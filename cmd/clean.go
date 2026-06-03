package cmd

import (
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClean()
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

func runClean() error {
	opts, cliOpts, err := prepareClean(cleanMinAge, cleanWorkers, cleanDryRun, cleanPermanent, cleanYes)
	if err != nil {
		return err
	}

	targets := cleaner.Registry()
	targets = cleaner.FilterTargets(targets, cleanCategory, cleanTargetIDs, cleanNoDeveloper, cleanNoBrowser)

	if len(targets) == 0 {
		fmt.Println("No cleanable targets found for the specified filters.")
		return nil
	}

	fmt.Println("Scanning...")
	opts.OnProgress = func(p cleaner.Progress) {
		fmt.Printf("\rScanning... %d/%d (%s)              ", p.Done, p.Total, p.Current)
	}
	result, err := cleaner.Scan(targets, opts)
	fmt.Println() // Clear the progress line
	if err != nil {
		return fmt.Errorf("error during scan: %w", err)
	}

	cleaner.RenderCLI(result, cliOpts)
	return nil
}

func prepareClean(minAgeStr string, workers int, dryRun, permanent, yes bool) (cleaner.ScanOptions, cleaner.CLIOptions, error) {
	opts := cleaner.ScanOptions{}
	cliOpts := cleaner.CLIOptions{
		DryRun:    dryRun,
		Permanent: permanent,
		Yes:       yes,
	}

	if minAgeStr != "" {
		duration, err := fsutil.ParseDuration(minAgeStr)
		if err != nil {
			return opts, cliOpts, fmt.Errorf("invalid value for --min-age: %w", err)
		}
		opts.MinAge = duration
	}

	if workers > 0 {
		opts.Concurrency = workers
	}
	return opts, cliOpts, nil
}
