package cmd

import (
	"fmt"
	"os"

	"dupclean/diskanalyzer"

	"github.com/spf13/cobra"
)

var (
	analyzeTopN           int
	analyzeDepth          int
	analyzeMinSize        int
	analyzeOlderThan      int
	analyzeByType         bool
	analyzeJSON           bool
	analyzeNoHidden       bool
	analyzeFollowSymlinks bool
	analyzeExclude        []string
	analyzeWorkers        int
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [folder]",
	Short: "Analyze disk usage",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runAnalyze(cmd, args[0])
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().IntVar(&analyzeTopN, "top", 20, "Show N largest files")
	analyzeCmd.Flags().IntVar(&analyzeDepth, "depth", 2, "Tree depth in CLI view")
	analyzeCmd.Flags().IntVar(&analyzeMinSize, "min-size", 0, "Exclude files smaller than MB megabytes")
	analyzeCmd.Flags().IntVar(&analyzeOlderThan, "older-than", 0, "Only include files not modified in N days")
	analyzeCmd.Flags().BoolVar(&analyzeByType, "by-type", false, "Show type breakdown instead of tree")
	analyzeCmd.Flags().BoolVar(&analyzeJSON, "json", false, "Output JSON to stdout")
	analyzeCmd.Flags().BoolVar(&analyzeNoHidden, "no-hidden", false, "Skip hidden files and folders")
	analyzeCmd.Flags().BoolVar(&analyzeFollowSymlinks, "follow-symlinks", false, "Follow symbolic links")
	analyzeCmd.Flags().StringSliceVar(&analyzeExclude, "exclude", []string{}, "Glob pattern to exclude (repeatable)")
	analyzeCmd.Flags().IntVar(&analyzeWorkers, "workers", 0, "Number of concurrent stat workers")
}

func runAnalyze(cmd *cobra.Command, root string) {
	opts := diskanalyzer.DefaultOptions()
	cliOpts := diskanalyzer.CLIOptions{
		Depth:     analyzeDepth,
		TopN:      analyzeTopN,
		ByType:    analyzeByType,
		OlderThan: analyzeOlderThan,
		MinSize:   int64(analyzeMinSize) * 1024 * 1024,
	}

	opts.IncludeHidden = !analyzeNoHidden
	opts.FollowSymlinks = analyzeFollowSymlinks
	opts.ExcludePaths = analyzeExclude
	if analyzeWorkers > 0 {
		opts.Concurrency = analyzeWorkers
	}

	// Validate path
	info, err := os.Stat(root)
	if err != nil {
		fmt.Printf("Error: cannot access '%s': %v\n", root, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("Error: '%s' is not a valid directory\n", root)
		os.Exit(1)
	}

	// Run analysis
	opts.Context = cmd.Context()
	result, errors, err := diskanalyzer.Walk(root, opts)
	if err != nil {
		fmt.Printf("Error: analysis failed: %v\n", err)
		os.Exit(1)
	}

	// Print non-fatal errors to stderr
	for _, e := range errors {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", e)
	}

	// Output
	if analyzeJSON {
		if err := diskanalyzer.ExportJSON(result, os.Stdout); err != nil {
			fmt.Printf("Error: JSON export failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		diskanalyzer.RenderCLI(result, cliOpts)
	}
}
