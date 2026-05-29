package gui

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"dupclean/internal/version"
	"dupclean/gui/components"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	githubRepo = "PopolQue/dupclean"
	githubAPI  = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type UpdaterState struct {
	Window          fyne.Window
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	IsChecking      bool
	IsDownloading   bool
	Progress        float64
}

func NewUpdaterState(w fyne.Window) *UpdaterState {
	return &UpdaterState{
		Window:         w,
		CurrentVersion: version.Version,
	}
}

func UpdaterWidget(state *UpdaterState) fyne.CanvasObject {
	versionLabel := widget.NewLabel(fmt.Sprintf("Current Version: %s", state.CurrentVersion))
	statusLabel := widget.NewLabel("Click the button below to check for updates.")

	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	var checkBtn *widget.Button
	checkBtn = widget.NewButtonWithIcon("Check for Updates", theme.ViewRefreshIcon(), func() {
		checkBtn.Disable()
		statusLabel.SetText("Checking GitHub for updates...")

		go func() {
			release, err := checkForUpdates()

			fyne.Do(func() {
				checkBtn.Enable()
				if err != nil {
					statusLabel.SetText(fmt.Sprintf("Error checking updates: %v", err))
					return
				}

				state.LatestVersion = release.TagName
				if isNewerVersion(state.CurrentVersion, state.LatestVersion) {
					state.UpdateAvailable = true
					statusLabel.SetText(fmt.Sprintf("New version available: %s", state.LatestVersion))
					showUpdateDialog(state, release)
				} else {
					statusLabel.SetText(fmt.Sprintf("You are up to date! (Latest: %s)", state.LatestVersion))
					dialog.ShowInformation("Up to Date", fmt.Sprintf("You are running the latest version (%s).", state.CurrentVersion), state.Window)
				}
			})
		}()
	})

	viewChangelogBtn := widget.NewButtonWithIcon("View Full Changelog", theme.InfoIcon(), func() {
		ShowFullChangelog(state.Window)
	})

	versionCard := widget.NewCard("Version Info", "Information about your current installation", versionLabel)
	statusCard := widget.NewCard("Update Status", "Current update progress and status", container.NewVBox(
		statusLabel,
		progressBar,
	))

	return components.ToolHome(
		"Check for Updates",
		"Keep DupClean up to date with the latest features",
		[]fyne.CanvasObject{versionCard},
		checkBtn,
		statusCard,
		container.NewHBox(layout.NewSpacer(), viewChangelogBtn, layout.NewSpacer()),
	)
}

func checkForUpdates() (*GitHubRelease, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(githubAPI)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func isNewerVersion(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	currParts := strings.Split(current, ".")
	lateParts := strings.Split(latest, ".")

	for i := 0; i < len(currParts) && i < len(lateParts); i++ {
		var c, l int
		_, _ = fmt.Sscanf(currParts[i], "%d", &c)
		_, _ = fmt.Sscanf(lateParts[i], "%d", &l)

		if l > c {
			return true
		}
		if c > l {
			return false
		}
	}

	return len(lateParts) > len(currParts)
}

func showUpdateDialog(state *UpdaterState, release *GitHubRelease) {
	titleText := widget.NewLabelWithStyle("Update Available", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	titleText.Importance = widget.HighImportance
	titleText.SizeName = theme.SizeNameSubHeadingText

	header := container.NewVBox(
		container.NewHBox(
			titleText,
			layout.NewSpacer(),
			canvas.NewText(release.TagName, theme.Color(theme.ColorNameForeground)),
		),
		widget.NewSeparator(),
	)

	// Filter body to only show highlights (exclude installation guide)
	body := release.Body
	if strings.Contains(body, "## Installation") {
		body = strings.Split(body, "## Installation")[0]
	}
	body = strings.TrimSpace(body)

	recentTitle := widget.NewLabelWithStyle("Release Highlights", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	recentContent := widget.NewRichTextFromMarkdown(body)
	recentContent.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		header,
		recentTitle,
		recentContent,
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	dialog.ShowCustomConfirm("New Update Found", "Download & Install", "Later", scroll, func(confirm bool) {
		if confirm {
			downloadAndInstallUpdate(state, release)
		}
	}, state.Window)
}

func downloadAndInstallUpdate(state *UpdaterState, release *GitHubRelease) {
	// Find correct asset for current platform
	var downloadURL string
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}

	// Match asset name like: dupclean-darwin-arm64.tar.gz
	pattern := fmt.Sprintf("dupclean-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext)
	for _, asset := range release.Assets {
		if asset.Name == pattern {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		dialog.ShowError(fmt.Errorf("could not find update asset for %s-%s\nExpected: %s\nAvailable assets: %d",
			runtime.GOOS, runtime.GOARCH, pattern, len(release.Assets)), state.Window)
		return
	}

	progressBar := widget.NewProgressBar()
	progressDialog := dialog.NewCustomWithoutButtons("Updating", container.NewVBox(
		widget.NewLabel("Downloading and installing update..."),
		progressBar,
	), state.Window)
	progressDialog.Show()

	go func() {
		err := performUpdate(downloadURL, progressBar.SetValue)

		fyne.Do(func() {
			progressDialog.Hide()
			if err != nil {
				dialog.ShowError(fmt.Errorf("update failed: %v", err), state.Window)
			} else {
				d := dialog.NewInformation("Update Complete", "DupClean has been updated and will now restart.", state.Window)
				d.SetOnClosed(func() {
					restartApp()
				})
				d.Show()
			}
		})
	}()
}

func restartApp() {
	executable, err := os.Executable()
	if err != nil {
		log.Printf("[Updater] Error getting executable path: %v", err)
		os.Exit(0)
	}

	log.Printf("[Updater] Restarting application: %s", executable)

	var cmd *exec.Cmd
	// On macOS, if we're in an .app bundle, use 'open -n' to restart properly
	if runtime.GOOS == "darwin" && strings.Contains(strings.ToLower(executable), ".app/") {
		appPath := executable[:strings.Index(strings.ToLower(executable), ".app/")+4]
		log.Printf("[Updater] macOS .app detected, using: open -n %s", appPath)
		cmd = exec.Command("open", "-n", appPath)
	} else {
		cmd = exec.Command(executable)
	}

	// Detach the process so it continues after we exit
	// On Windows, Start() without waiting is enough.
	// On Unix, the process continues even if the parent exits.
	err = cmd.Start()

	if err != nil {
		log.Printf("[Updater] Failed to restart application: %v", err)
	}

	os.Exit(0)
}

func performUpdate(url string, setProgress func(float64)) error {
	// 1. Download binary to temp file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	tmpFile, err := os.CreateTemp("", "dupclean-update-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	defer func() { _ = tmpFile.Close() }()

	// Track download progress
	size := resp.ContentLength
	var downloaded int64
	buffer := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, werr := tmpFile.Write(buffer[:n]); werr != nil {
				return werr
			}
			downloaded += int64(n)
			if size > 0 {
				setProgress(float64(downloaded) / float64(size) * 0.4) // Download is first 40%
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	_ = tmpFile.Close()

	// 2. Download and verify checksum
	setProgress(0.5)
	checksumURL := url + ".sha256"
	checksumResp, err := http.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksum: %v", err)
	}
	defer func() { _ = checksumResp.Body.Close() }()

	checksumBytes, err := io.ReadAll(checksumResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read checksum: %v", err)
	}
	expectedHash := strings.TrimSpace(string(checksumBytes))

	isValid, err := verifyHash(tmpFile.Name(), expectedHash)
	if err != nil {
		return fmt.Errorf("failed to verify hash: %v", err)
	}
	if !isValid {
		return fmt.Errorf("checksum verification failed! The file may be tampered with.")
	}

	// 3. Extract binary to a temporary location first
	setProgress(0.6)
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// On macOS, if we are inside an .app bundle, executable is something like
	// /Applications/DupClean.app/Contents/MacOS/dupclean
	// We extract to a truly temporary file to avoid "permission denied" during extraction.
	tmpNewBin, err := os.CreateTemp("", "dupclean-new-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file for extraction: %v", err)
	}
	newBinaryPath := tmpNewBin.Name()
	_ = tmpNewBin.Close()
	defer func() { _ = os.Remove(newBinaryPath) }()

	if strings.HasSuffix(url, ".zip") {
		err = extractFromZip(tmpFile.Name(), newBinaryPath)
	} else {
		err = extractFromTarGz(tmpFile.Name(), newBinaryPath)
	}
	if err != nil {
		return err
	}

	// 4. Replace current binary
	setProgress(0.9)

	// Attempt the move. If it fails with permission denied on macOS, we might need elevation.
	installNewBinary := func(src, dst string) error {
		// On Windows, we rename the current binary to .old and move the new one in
		if runtime.GOOS == "windows" {
			oldPath := dst + ".old"
			_ = os.Remove(oldPath) // Remove previous old if exists
			if err := os.Rename(dst, oldPath); err != nil {
				return fmt.Errorf("failed to rename current binary: %v", err)
			}
		} else {
			// On Unix, try to remove the old one first
			_ = os.Remove(dst)
		}

		if err := os.Rename(src, dst); err != nil {
			// If rename fails (e.g. cross-device or permission), try to copy
			if copyErr := copyFile(src, dst); copyErr != nil {
				return fmt.Errorf("failed to install new binary: %v (rename error: %v)", copyErr, err)
			}
		}
		return nil
	}

	err = installNewBinary(newBinaryPath, executable)
	if err != nil {
		// If it's a permission error on macOS, try elevation via osascript
		if strings.Contains(err.Error(), "permission denied") && runtime.GOOS == "darwin" {
			if elevErr := macOSInstallWithElevation(newBinaryPath, executable); elevErr == nil {
				setProgress(1.0)
				return nil
			}
		}

		// If it's a permission error on macOS/Linux, give better instructions
		if strings.Contains(err.Error(), "permission denied") {
			if runtime.GOOS == "darwin" && strings.HasPrefix(executable, "/Applications/") {
				return fmt.Errorf("permission denied (try running: sudo xattr -rd com.apple.quarantine %s && brew install PopolQue/dupclean/dupclean)", executable)
			}
			return fmt.Errorf("%v (you may need administrative privileges)", err)
		}
		return err
	}

	// Set permissions on Unix
	if runtime.GOOS != "windows" {
		_ = os.Chmod(executable, 0755)
	}

	setProgress(1.0)
	return nil
}

// macOSInstallWithElevation uses osascript to request administrator privileges for the file move.
// It performs strict path validation to prevent privilege escalation.
func macOSInstallWithElevation(src, dst string) error {
	// Validate destination is in a trusted location
	trustedPaths := []string{"/Applications/", "/usr/local/bin/"}
	isTrusted := false
	for _, p := range trustedPaths {
		if strings.HasPrefix(dst, p) {
			isTrusted = true
			break
		}
	}
	if !isTrusted {
		return fmt.Errorf("refusing to elevate: destination path %s is not in a trusted location", dst)
	}

	// do shell script in AppleScript uses /bin/sh. We use AppleScript's 'quoted form of' to handle spaces safely.
	// The AppleScript command looks like:
	// do shell script "cp -f " & quoted form of "/src" & " " & quoted form of "/dst" & " && chmod 755 " & quoted form of "/dst" with administrator privileges
	asCommand := fmt.Sprintf("do shell script \"cp -f \" & quoted form of %q & \" \" & quoted form of %q & \" && chmod 755 \" & quoted form of %q with administrator privileges",
		src, dst, dst)

	cmd := exec.Command("osascript", "-e", asCommand)
	return cmd.Run()
}

// verifyHash calculates the SHA-256 hash of the file and compares it to the expected hash.
func verifyHash(filePath string, expectedHash string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}

	actualHash := hex.EncodeToString(h.Sum(nil))
	return strings.EqualFold(actualHash, expectedHash), nil
}

// copyFile is a fallback for os.Rename when moving across filesystems or when rename fails
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()

	// Try to open destination for writing. This will still fail if permission denied.
	destination, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() { _ = destination.Close() }()

	_, err = io.Copy(destination, source)
	return err
}

func extractFromTarGz(tarGzPath, destPath string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Find the binary in the tarball (it's named dupclean-...)
		if header.Typeflag == tar.TypeReg && strings.Contains(strings.ToLower(header.Name), "dupclean") {
			out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				_ = out.Close()
				return err
			}
			_ = out.Close()
			return nil
		}
	}
	return fmt.Errorf("binary not found in update archive")
}

func extractFromZip(zipPath, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		if !f.FileInfo().IsDir() && strings.Contains(strings.ToLower(f.Name), "dupclean") {
			rc, err := f.Open()
			if err != nil {
				return err
			}

			out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				_ = rc.Close()
				return err
			}

			if _, err := io.Copy(out, rc); err != nil {
				_ = out.Close()
				_ = rc.Close()
				return err
			}

			_ = out.Close()
			_ = rc.Close()
			return nil
		}
	}
	return fmt.Errorf("binary not found in update archive")
}
