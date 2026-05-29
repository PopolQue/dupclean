package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"dupclean/internal/fsutil"
	"dupclean/internal/trash"
	"dupclean/scanner"
)

var (
	stdin  io.Reader = os.Stdin
	stdout io.Writer = os.Stdout
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
		fmt.Fprintf(stdout, "\n%s%s No duplicates found!%s Your drive is clean.%s\n\n", colorBold, colorGreen, colorReset, colorReset)
		fmt.Fprintf(stdout, "%s Scan completed in %s%s — %d files checked%s\n\n",
			colorDim, stats.ScanDuration.Round(time.Second), colorReset, stats.TotalScanned, colorReset)
		return
	}

	// Sort groups by wasted bytes (biggest offenders first)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Files[0].Size > groups[j].Files[0].Size
	})

	printScanSummary(stats, len(groups))
	printControlsHelp()

	reader := bufio.NewReader(stdin)
	deletedCount := 0
	var freedBytes int64

	for i, group := range groups {
		fmt.Fprintf(stdout, "\n%s%s", colorCyan, strings.Repeat("─", 70))
		fmt.Fprintf(stdout, "%s\n", colorReset)
		fmt.Fprintf(stdout, "%s Group %d of %d%s%s • identical audio content • %s%s each%s\n",
			colorBold+colorWhite, i+1, len(groups), colorReset,
			colorGray, colorDim, fsutil.FormatBytes(group.Files[0].Size), colorReset)
		fmt.Fprintf(stdout, "%s%s%s\n", colorCyan, strings.Repeat("─", 70), colorReset)

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
			fmt.Fprintf(stdout, "\n  %s  %s%s%s\n", num, colorBold, f.Name, colorReset)
			fmt.Fprintf(stdout, "       %s%s%s\n", colorGray, f.Path, colorReset)
			fmt.Fprintf(stdout, "       %s %s  •  %s%s\n",
				colorDim, fsutil.FormatBytes(f.Size), f.ModTime.Format("2006-01-02 15:04"), colorReset)
		}

		fmt.Fprintf(stdout, "\n%s  Keep which file?%s\n", colorBold, colorReset)
		fmt.Fprintf(stdout, " %s[1-%d]%s Keep that file, delete others\n", colorYellow, len(files), colorReset)
		fmt.Fprintf(stdout, " %s[s]%s Skip this group\n", colorYellow, colorReset)
		fmt.Fprintf(stdout, " %s[a]%s Skip all remaining groups\n", colorYellow, colorReset)
		fmt.Fprintf(stdout, " %s[q]%s Quit\n", colorYellow, colorReset)
		fmt.Fprintf(stdout, "\n %s>%s ", colorCyan, colorReset)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "q", "quit":
			fmt.Fprintf(stdout, "\n%s⏹ Stopped early.%s You can resume later.\n\n", colorYellow+colorBold, colorReset)
			printFinalSummary(deletedCount, freedBytes)
			return
		case "s", "skip", "":
			fmt.Fprintf(stdout, "  %s↷ Skipped this group%s\n", colorGray, colorReset)
			continue
		case "a":
			fmt.Fprintf(stdout, "\n%s↷ Skipping all remaining groups.%s\n", colorGray+colorBold, colorReset)
			printFinalSummary(deletedCount, freedBytes)
			return
		default:
			choice, err := strconv.Atoi(input)
			if err != nil || choice < 1 || choice > len(files) {
				fmt.Fprintf(stdout, " %s Invalid choice. Please enter a number between 1 and %d.%s\n", colorYellow, len(files), colorReset)
				continue
			}

			keepFile := files[choice-1]
			fmt.Fprintf(stdout, "\n  %s ✓ Keeping:%s %s%s%s\n", colorGreen+colorBold, colorReset, colorWhite, keepFile.Name, colorReset)

			for idx, f := range files {
				if idx == choice-1 {
					continue
				}
				if err := moveToTrash(f.Path); err != nil {
					fmt.Fprintf(stdout, " %s Could not delete %s: %v%s\n", colorRed, f.Name, err, colorReset)
				} else {
					fmt.Fprintf(stdout, " %s Deleted:%s %s%s%s\n", colorRed, colorReset, colorGray, f.Name, colorReset)
					deletedCount++
					freedBytes += f.Size
				}
			}
		}
	}

	printFinalSummary(deletedCount, freedBytes)
}

// moveToTrash uses the unified trash package for cross-platform trash support
func moveToTrash(path string) error {
	return trash.MoveToTrash(path)
}

func printHeader() {
	fmt.Fprint(stdout, colorReset)
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "%s╔═══════════════════════════════════════════════════════╗%s\n", colorPurple+colorBold, colorReset)
	fmt.Fprintf(stdout, "%s║%s          %sDUPCLEAN%s  — Duplicate File Hunter            %s║%s\n",
		colorPurple, colorReset, colorBold+colorWhite, colorReset, colorPurple, colorReset)
	fmt.Fprintf(stdout, "%s╚═══════════════════════════════════════════════════════╝%s\n\n", colorPurple+colorBold, colorReset)
}

func printScanSummary(stats scanner.ScanStats, groupCount int) {
	fmt.Fprintf(stdout, "%sScan Summary%s\n", colorBold+colorCyan, colorReset)
	fmt.Fprintf(stdout, "%s\n", strings.Repeat("─", 40))
	fmt.Fprintf(stdout, "   %sDuration:%s    %s%s%s\n", colorGray, colorReset, colorWhite, stats.ScanDuration.Round(time.Second), colorReset)
	fmt.Fprintf(stdout, "   %sFiles:%s       %s%d%s\n", colorGray, colorReset, colorWhite, stats.TotalScanned, colorReset)
	fmt.Fprintf(stdout, "   %sGroups:%s      %s%d%s\n", colorGray, colorReset, colorWhite, groupCount, colorReset)
	fmt.Fprintf(stdout, "   %sExtra:%s       %s%d%s copies\n", colorGray, colorReset, colorWhite, stats.TotalDupes, colorReset)
	fmt.Fprintf(stdout, "   %sWasted:%s      %s%s%s\n", colorGray, colorReset, colorRed+colorBold, fsutil.FormatBytes(stats.WastedBytes), colorReset)
	fmt.Fprintln(stdout)
}

func printControlsHelp() {
	fmt.Fprintf(stdout, "%sControls:%s\n", colorBold+colorUnderline, colorReset)
	fmt.Fprintf(stdout, "  %s[1-9]%s  Keep file #, delete others\n", colorYellow, colorReset)
	fmt.Fprintf(stdout, "  %s[s]%s    Skip this group\n", colorYellow, colorReset)
	fmt.Fprintf(stdout, "  %s[a]%s    Skip all remaining\n", colorYellow, colorReset)
	fmt.Fprintf(stdout, "  %s[q]%s    Quit\n", colorYellow, colorReset)
	fmt.Fprintln(stdout)
}

func printFinalSummary(deleted int, freed int64) {
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "%s%s%s\n", colorCyan, strings.Repeat("─", 70), colorReset)

	if deleted == 0 {
		fmt.Fprintf(stdout, "\n%s Nothing was deleted.%s Your files are safe.\n\n", colorYellow+colorBold, colorReset)
	} else {
		fmt.Fprintf(stdout, "\n  %s Cleanup Complete!%s\n\n", colorGreen+colorBold, colorReset)
		fmt.Fprintf(stdout, "      Files deleted:  %s%d%s\n", colorBold, deleted, colorReset)
		fmt.Fprintf(stdout, "      Space freed:    %s%s%s\n\n", colorGreen+colorBold, fsutil.FormatBytes(freed), colorReset)
		fmt.Fprintf(stdout, "  %s Tip: Empty your Trash to reclaim disk space.%s\n\n", colorDim, colorReset)
	}

	fmt.Fprintf(stdout, "%s%s%s\n", colorCyan, strings.Repeat("─", 70), colorReset)
	fmt.Fprintln(stdout)
}
