package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"dupclean/internal/scanner"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// Run is the main entry point for the interactive UI
func Run(groups []scanner.DuplicateGroup, stats scanner.ScanStats) {
	printHeader()

	if len(groups) == 0 {
		fmt.Printf("\n%s✅ No duplicates found!%s Your drive is clean.\n\n", colorGreen+colorBold, colorReset)
		fmt.Printf("%sScan completed in %s — %d files checked%s\n\n",
			colorGray, stats.ScanDuration.Round(1000000), stats.TotalScanned, colorReset)
		return
	}

	// Sort groups by wasted bytes (biggest offenders first)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Files[0].Size > groups[j].Files[0].Size
	})

	printScanSummary(stats, len(groups))

	reader := bufio.NewReader(os.Stdin)
	deletedCount := 0
	var freedBytes int64

	for i, group := range groups {
		fmt.Printf("\n%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorGray, colorReset)
		fmt.Printf("%s Group %d of %d%s  %s(identical audio content)%s\n",
			colorBold+colorCyan, i+1, len(groups), colorReset,
			colorGray, colorReset)
		fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorGray, colorReset)

		// Sort files: prefer files higher in the directory tree (shorter path)
		files := group.Files
		sort.Slice(files, func(i, j int) bool {
			di := strings.Count(files[i].Path, string(os.PathSeparator))
			dj := strings.Count(files[j].Path, string(os.PathSeparator))
			if di != dj {
				return di < dj
			}
			return files[i].ModTime.Before(files[j].ModTime)
		})

		for idx, f := range files {
			num := fmt.Sprintf("[%d]", idx+1)
			fmt.Printf("\n  %s%s%s  %s%s%s\n",
				colorYellow+colorBold, num, colorReset,
				colorWhite+colorBold, f.Name, colorReset)
			fmt.Printf("       %s%s%s\n", colorGray, f.Path, colorReset)
			fmt.Printf("       %sSize: %s  Modified: %s%s\n",
				colorDim, formatBytes(f.Size), f.ModTime.Format("2006-01-02 15:04"), colorReset)
		}

		fmt.Printf("\n%s  Keep which file? (1-%d)  [s]kip  [a]ll skip  [q]uit%s\n  > ",
			colorGray, len(files), colorReset)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "q", "quit":
			fmt.Printf("\n%s⏹  Stopped early.%s\n", colorYellow, colorReset)
			goto done
		case "s", "skip", "":
			fmt.Printf("  %s↷ Skipped%s\n", colorGray, colorReset)
			continue
		case "a":
			fmt.Printf("\n%s↷ Skipping all remaining groups.%s\n", colorGray, colorReset)
			goto done
		default:
			choice, err := strconv.Atoi(input)
			if err != nil || choice < 1 || choice > len(files) {
				fmt.Printf("  %s⚠️  Invalid choice, skipping.%s\n", colorYellow, colorReset)
				continue
			}

			keepFile := files[choice-1]
			for idx, f := range files {
				if idx == choice-1 {
					fmt.Printf("  %s✓ Keeping:%s %s\n", colorGreen, colorReset, f.Name)
					continue
				}
				if err := moveToTrash(f.Path); err != nil {
					fmt.Printf("  %s❌ Could not trash %s: %v%s\n", colorRed, f.Name, err, colorReset)
				} else {
					fmt.Printf("  %s🗑  Trashed:%s %s%s%s\n", colorRed, colorReset, colorGray, f.Name, colorReset)
					deletedCount++
					freedBytes += f.Size
					_ = keepFile
				}
			}
		}
	}

done:
	printFinalSummary(deletedCount, freedBytes)
}

// moveToTrash uses macOS `trash` command or AppleScript to move a file to Trash
func moveToTrash(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Try using the `trash` CLI tool first (brew install trash)
	if _, err := exec.LookPath("trash"); err == nil {
		return exec.Command("trash", absPath).Run()
	}

	// Fall back to AppleScript (built-in on macOS)
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, absPath)
	return exec.Command("osascript", "-e", script).Run()
}

func printHeader() {
	fmt.Print(colorReset)
	fmt.Println()
	fmt.Printf("%s╔═══════════════════════════════════════╗%s\n", colorPurple, colorReset)
	fmt.Printf("%s║%s  %s🎧 DUPCLEAN%s  — Audio Duplicate Hunter  %s║%s\n",
		colorPurple, colorReset, colorBold+colorWhite, colorReset, colorPurple, colorReset)
	fmt.Printf("%s╚═══════════════════════════════════════╝%s\n\n", colorPurple, colorReset)
}

func printScanSummary(stats scanner.ScanStats, groupCount int) {
	fmt.Printf("%s📊 Scan complete in %s%s\n", colorBold, stats.ScanDuration.Round(1000000), colorReset)
	fmt.Printf("   %sFiles scanned:%s   %d\n", colorGray, colorReset, stats.TotalScanned)
	fmt.Printf("   %sDuplicate groups:%s %d\n", colorGray, colorReset, groupCount)
	fmt.Printf("   %sExtra copies:%s    %d  (%s wasted)\n",
		colorGray, colorReset, stats.TotalDupes, formatBytes(stats.WastedBytes))
}

func printFinalSummary(deleted int, freed int64) {
	fmt.Println()
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorGray, colorReset)
	if deleted == 0 {
		fmt.Printf("\n%s  Nothing was deleted. Your files are safe.%s\n\n", colorYellow, colorReset)
	} else {
		fmt.Printf("\n  %s✅ Done!%s  Moved %s%d file(s)%s to Trash  →  freed %s%s%s\n\n",
			colorGreen+colorBold, colorReset,
			colorBold, deleted, colorReset,
			colorBold, formatBytes(freed), colorReset)
		fmt.Printf("  %sTip: Empty your Trash to reclaim disk space.%s\n\n", colorGray, colorReset)
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
