package gui

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"dupclean/internal/version"

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
	title := canvas.NewText("Check for Updates", theme.Color(theme.ColorNamePrimary))
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}

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

	content := container.NewVBox(
		title,
		versionLabel,
		layout.NewSpacer(),
		statusLabel,
		progressBar,
		checkBtn,
	)

	return container.NewCenter(content)
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
	info := widget.NewLabel(fmt.Sprintf("A new version (%s) is available.\n\nChanges:\n%s", release.TagName, release.Body))

	dialog.ShowCustomConfirm("Update Available", "Download & Install", "Later", container.NewVScroll(info), func(confirm bool) {
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
				dialog.ShowInformation("Update Complete", "DupClean has been updated. Please restart the application.", state.Window)
			}
		})
	}()
}

func performUpdate(url string, setProgress func(float64)) error {
	// 1. Download to temp file
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
				setProgress(float64(downloaded) / float64(size) * 0.5) // Download is first 50%
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

	// 2. Extract binary
	setProgress(0.6)
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	newBinaryPath := executable + ".new"

	if strings.HasSuffix(url, ".zip") {
		err = extractFromZip(tmpFile.Name(), newBinaryPath)
	} else {
		err = extractFromTarGz(tmpFile.Name(), newBinaryPath)
	}
	if err != nil {
		return err
	}

	// 3. Replace current binary
	setProgress(0.9)

	// On Unix, we can rename over the running binary
	// On Windows, we rename the current binary to .old and move the new one in
	if runtime.GOOS == "windows" {
		oldPath := executable + ".old"
		_ = os.Remove(oldPath) // Remove previous old if exists
		if err := os.Rename(executable, oldPath); err != nil {
			return fmt.Errorf("failed to rename current binary: %v", err)
		}
	} else {
		// On Unix, remove the old one first if it exists (might be busy)
		_ = os.Remove(executable)
	}

	if err := os.Rename(newBinaryPath, executable); err != nil {
		return fmt.Errorf("failed to install new binary: %v", err)
	}

	// Set permissions on Unix
	if runtime.GOOS != "windows" {
		_ = os.Chmod(executable, 0755)
	}

	setProgress(1.0)
	return nil
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
		if header.Typeflag == tar.TypeReg && strings.Contains(header.Name, "dupclean") {
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
		if !f.FileInfo().IsDir() && strings.Contains(f.Name, "dupclean") {
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
