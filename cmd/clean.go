package cmd

import (
	"fmt"
	"os"

	"github.com/PopolQue/dupclean/cleaner"
	"github.com/PopolQue/dupclean/internal/fsutil"

	"github.com/spf13/cobra"
)

var (
	cleanCategory    string
	cleanTargetIDs   []string
	cleanMinAge      string
	cleanPermanent   bool
	cleanDryRun      bool
	cleanYes         bool
	cleanWorkers     int
	cleanNoDeveloper bool
	cleanNoBrowser   bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Cleanup cache & temp files",
	Run: func(cmd *cobra.Command, args []string) {
		runClean()
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().StringVar(&cleanCategory, "category", "", "Only scan targets in this category (system|browser|developer|logs)")
	cleanCmd.Flags().StringSliceVar(&cleanTargetIDs, "target", []string{}, "Only scan this specific target ID (repeatable)")
	cleanCmd.Flags().StringVar(&cleanMinAge, "min-age", "", "Only include files older than this (e.g. 24h, 7d)")
	cleanCmd.Flags().BoolVar(&cleanPermanent, "permanent", false, "Delete permanently instead of moving to Trash")
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "Show what would be deleted without deleting anything")
	cleanCmd.Flags().BoolVar(&cleanYes, "yes", false, "Skip interactive confirmation")
	cleanCmd.Flags().IntVar(&cleanWorkers, "workers", 0, "Number of concurrent scan workers")
	cleanCmd.Flags().BoolVar(&cleanNoDeveloper, "no-developer", false, "Exclude developer targets")
	cleanCmd.Flags().BoolVar(&cleanNoBrowser, "no-browser", false, "Exclude browser targets")
}

func runClean() {
	opts := cleaner.ScanOptions{}
	cliOpts := cleaner.CLIOptions{
		DryRun:    cleanDryRun,
		Permanent: cleanPermanent,
		Yes:       cleanYes,
	}

	if cleanMinAge != "" {
		duration, err := fsutil.ParseDuration(cleanMinAge)
		if err != nil {
			fmt.Printf("Error: invalid value for --min-age: %v\n", err)
			os.Exit(1)
		}
		opts.MinAge = duration
	}

	if cleanWorkers > 0 {
		opts.Concurrency = cleanWorkers
	}

	targets := cleaner.Registry()
	targets = cleaner.FilterTargets(targets, cleanCategory, cleanTargetIDs, cleanNoDeveloper, cleanNoBrowser)

	if len(targets) == 0 {
		fmt.Println("No cleanable targets found for the specified filters.")
		return
	}

	fmt.Println("Scanning...")
	result, err := cleaner.Scan(targets, opts)
	if err != nil {
		fmt.Printf("Error during scan: %v\n", err)
		os.Exit(1)
	}

	cleaner.RenderCLI(result, cliOpts)
}
