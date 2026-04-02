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
	"time"

	"dupclean/scanner"
)

const (
	colorReset     = "\033[0m"
	colorRed       = "\033[31m"
	colorGreen     = "\033[32m"
	colorYellow    = "\033[33m"
	colorBlue      = "\033[34m"
	colorPurple    = "\033[35m"
	colorCyan      = "\033[36m"
	colorWhite     = "\033[37m"
	colorGray      = "\033[90m"
	colorBold      = "\033[1m"
	colorDim       = "\033[2m"
	colorUnderline = "\033[4m"
)

// Run is the main entry point for the interactive UI
func Run(groups []scanner.DuplicateGroup, stats scanner.ScanStats) {
	printHeader()

	if len(groups) == 0 {
		fmt.Printf("\n%s%s No duplicates found!%s Your drive is clean.%s\n\n", colorBold, colorGreen, colorReset, colorReset)
		fmt.Printf("%s Scan completed in %s%s — %d files checked%s\n\n",
			colorDim, stats.ScanDuration.Round(time.Second), colorReset, stats.TotalScanned, colorReset)
		return
	}

	// Sort groups by wasted bytes (biggest offenders first)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Files[0].Size > groups[j].Files[0].Size
	})

	printScanSummary(stats, len(groups))
	printControlsHelp()

	reader := bufio.NewReader(os.Stdin)
	deletedCount := 0
	var freedBytes int64

	for i, group := range groups {
		fmt.Printf("\n%s%s", colorCyan, strings.Repeat("─", 70))
		fmt.Printf("%s\n", colorReset)
		fmt.Printf("%s Group %d of %d%s%s • identical audio content • %s%s each%s\n",
			colorBold+colorWhite, i+1, len(groups), colorReset,
			colorGray, colorDim, formatBytes(group.Files[0].Size), colorReset)
		fmt.Printf("%s%s%s\n", colorCyan, strings.Repeat("─", 70), colorReset)

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
			num := fmt.Sprintf("%s[%d]%s", colorYellow+colorBold, idx+1, colorReset)
			fmt.Printf("\n  %s  %s%s%s\n", num, colorBold, f.Name, colorReset)
			fmt.Printf("       %s%s%s\n", colorGray, f.Path, colorReset)
			fmt.Printf("       %s %s  •  %s%s\n",
				colorDim, formatBytes(f.Size), f.ModTime.Format("2006-01-02 15:04"), colorReset)
		}

		fmt.Printf("\n%s  Keep which file?%s\n", colorBold, colorReset)
		fmt.Printf(" %s[1-%d]%s Keep that file, delete others\n", colorYellow, len(files), colorReset)
		fmt.Printf(" %s[s]%s Skip this group\n", colorYellow, colorReset)
		fmt.Printf(" %s[a]%s Skip all remaining groups\n", colorYellow, colorReset)
		fmt.Printf(" %s[q]%s Quit\n", colorYellow, colorReset)
		fmt.Printf("\n %s>%s ", colorCyan, colorReset)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "q", "quit":
			fmt.Printf("\n%s⏹ Stopped early.%s You can resume later.\n\n", colorYellow+colorBold, colorReset)
			printFinalSummary(deletedCount, freedBytes)
			return
		case "s", "skip", "":
			fmt.Printf("  %s↷ Skipped this group%s\n", colorGray, colorReset)
			continue
		case "a":
			fmt.Printf("\n%s↷ Skipping all remaining groups.%s\n", colorGray+colorBold, colorReset)
			printFinalSummary(deletedCount, freedBytes)
			return
		default:
			choice, err := strconv.Atoi(input)
			if err != nil || choice < 1 || choice > len(files) {
				fmt.Printf(" %s Invalid choice. Please enter a number between 1 and %d.%s\n", colorYellow, len(files), colorReset)
				continue
			}

			keepFile := files[choice-1]
			fmt.Printf("\n  %s ✓ Keeping:%s %s%s%s\n", colorGreen+colorBold, colorReset, colorWhite, keepFile.Name, colorReset)

			for idx, f := range files {
				if idx == choice-1 {
					continue
				}
				if err := moveToTrash(f.Path); err != nil {
					fmt.Printf(" %s Could not delete %s: %v%s\n", colorRed, f.Name, err, colorReset)
				} else {
					fmt.Printf(" %s Deleted:%s %s%s%s\n", colorRed, colorReset, colorGray, f.Name, colorReset)
					deletedCount++
					freedBytes += f.Size
				}
			}
		}
	}

	printFinalSummary(deletedCount, freedBytes)
}

// moveToTrash uses the cleaner package's secure trash mechanism
func moveToTrash(path string) error {
	if path == "" {
		return fmt.Errorf("cannot move empty path to trash")
	}
	
	// Import locally to avoid circular dependency
	// We inline the safe implementation here
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Verify path exists
	if _, err := os.Stat(absPath); err != nil {
		return err
	}

	// Try using the `trash` CLI tool first
	if _, err := exec.LookPath("trash"); err == nil {
		return exec.Command("trash", absPath).Run()
	}

	// Fall back to AppleScript with proper escaping
	escapedPath := strings.ReplaceAll(absPath, "\\", "\\\\")
	escapedPath = strings.ReplaceAll(escapedPath, "\"", "\\\"")
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, escapedPath)
	return exec.Command("osascript", "-e", script).Run()
}

func printHeader() {
	fmt.Print(colorReset)
	fmt.Println()
	fmt.Printf("%s╔═══════════════════════════════════════════════════════╗%s\n", colorPurple+colorBold, colorReset)
	fmt.Printf("%s║%s          %sDUPCLEAN%s  — Duplicate File Hunter            %s║%s\n",
		colorPurple, colorReset, colorBold+colorWhite, colorReset, colorPurple, colorReset)
	fmt.Printf("%s╚═══════════════════════════════════════════════════════╝%s\n\n", colorPurple+colorBold, colorReset)
}

func printScanSummary(stats scanner.ScanStats, groupCount int) {
	fmt.Printf("%sScan Summary%s\n", colorBold+colorCyan, colorReset)
	fmt.Printf("%s\n", strings.Repeat("─", 40))
	fmt.Printf("   %sDuration:%s     %s%s%s\n", colorGray, colorReset, colorWhite, stats.ScanDuration.Round(time.Second), colorReset)
	fmt.Printf("   %sFiles:%s       %s%d%s\n", colorGray, colorReset, colorWhite, stats.TotalScanned, colorReset)
	fmt.Printf("   %sGroups:%s      %s%d%s\n", colorGray, colorReset, colorWhite, groupCount, colorReset)
	fmt.Printf("   %sExtra:%s       %s%d%s copies\n", colorGray, colorReset, colorWhite, stats.TotalDupes, colorReset)
	fmt.Printf("   %sWasted:%s      %s%s%s\n", colorGray, colorReset, colorRed+colorBold, formatBytes(stats.WastedBytes), colorReset)
	fmt.Println()
}

func printControlsHelp() {
	fmt.Printf("%sControls:%s\n", colorBold+colorUnderline, colorReset)
	fmt.Printf("  %s[1-9]%s  Keep file #, delete others\n", colorYellow, colorReset)
	fmt.Printf("  %s[s]%s    Skip this group\n", colorYellow, colorReset)
	fmt.Printf("  %s[a]%s    Skip all remaining\n", colorYellow, colorReset)
	fmt.Printf("  %s[q]%s    Quit\n", colorYellow, colorReset)
	fmt.Println()
}

func printFinalSummary(deleted int, freed int64) {
	fmt.Println()
	fmt.Printf("%s%s%s\n", colorCyan, strings.Repeat("─", 70), colorReset)

	if deleted == 0 {
		fmt.Printf("\n%s Nothing was deleted.%s Your files are safe.\n\n", colorYellow+colorBold, colorReset)
	} else {
		fmt.Printf("\n  %s Cleanup Complete!%s\n\n", colorGreen+colorBold, colorReset)
		fmt.Printf("      Files deleted:  %s%d%s\n", colorBold, deleted, colorReset)
		fmt.Printf("      Space freed:    %s%s%s\n\n", colorGreen+colorBold, formatBytes(freed), colorReset)
		fmt.Printf("  %s Tip: Empty your Trash to reclaim disk space.%s\n\n", colorDim, colorReset)
	}

	fmt.Printf("%s%s%s\n", colorCyan, strings.Repeat("─", 70), colorReset)
	fmt.Println()
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
