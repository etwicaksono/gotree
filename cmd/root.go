package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/etwicaksono/gotree/internal/config"
	"github.com/etwicaksono/gotree/internal/filter"
	"github.com/etwicaksono/gotree/internal/format"
	"github.com/etwicaksono/gotree/internal/output"
	"github.com/etwicaksono/gotree/internal/version"
	"github.com/etwicaksono/gotree/internal/walker"
)

var (
	rootPath        string
	outputPath      string
	excludePatterns []string
	excludeDirs     []string
	excludeFiles    []string
	maxDepth        int
	useGitignore    bool
	followSymlinks  bool
	excludeHidden   bool
)

var rootCmd = &cobra.Command{
	Use:   "gotree",
	Short: "Generate a markdown-formatted directory tree",
	Long:  "gotree generates a markdown-formatted directory and file tree with glob-based exclusion.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate root path
		rp := rootPath
		if rp == "" {
			rp = "."
		}
		rp = filepath.Clean(rp)
		if _, err := os.Stat(rp); err != nil {
			return fmt.Errorf("root path: %w", err)
		}

		// Build config
		cfg := config.Config{
			Root:           rp,
			Output:         outputPath,
			Exclude:        excludePatterns,
			ExcludeDir:     excludeDirs,
			ExcludeFile:    excludeFiles,
			MaxDepth:       maxDepth,
			UseGitignore:   useGitignore,
			FollowSymlinks: followSymlinks,
			ExcludeHidden:  excludeHidden,
		}

		// Prepare filter (optional gitignore)
		filt, err := filter.New(rp, cfg)
		if err != nil {
			return fmt.Errorf("init filter: %w", err)
		}

		// Walk filesystem
		tree, counts, err := walker.Walk(rp, cfg, filt)
		if err != nil {
			return fmt.Errorf("walk: %w", err)
		}

		// Render Markdown
		md := format.RenderMarkdown(rp, cfg, tree, counts)

		// Write output
		if err := output.WriteOutput(outputPath, md); err != nil {
			return fmt.Errorf("write output: %w", err)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version.Version

	rootCmd.Flags().StringVar(&rootPath, "root", ".", "Root path to traverse")
	rootCmd.Flags().StringVar(&outputPath, "output", "./TREE.md", "Markdown output file path")
	rootCmd.Flags().StringArrayVar(&excludePatterns, "exclude", nil, "Glob patterns applied to names (repeatable)")
	rootCmd.Flags().StringArrayVar(&excludeDirs, "exclude-dir", nil, "Glob patterns applied to directories (repeatable)")
	rootCmd.Flags().StringArrayVar(&excludeFiles, "exclude-file", nil, "Glob patterns applied to files (repeatable)")
	rootCmd.Flags().IntVar(&maxDepth, "max-depth", 0, "Maximum depth to traverse (0 = unlimited)")
	rootCmd.Flags().BoolVar(&useGitignore, "gitignore", false, "Apply rules from .gitignore at root")
	rootCmd.Flags().BoolVar(&followSymlinks, "follow-symlinks", false, "Traverse directory symlinks")
	rootCmd.Flags().BoolVar(&excludeHidden, "exclude-hidden", false, "Exclude hidden files and directories")
}
