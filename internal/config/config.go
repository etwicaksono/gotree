package config

// Config holds runtime options for generating the directory tree.
type Config struct {
	Root           string   // starting path for traversal
	Output         string   // markdown output path
	Exclude        []string // glob patterns applied to both files and directories
	ExcludeDir     []string // glob patterns applied only to directories
	ExcludeFile    []string // glob patterns applied only to files
	MaxDepth       int      // 0 = unlimited
	UseGitignore   bool     // apply .gitignore rules at root
	FollowSymlinks bool     // traverse directory symlinks
	ExcludeHidden  bool     // exclude hidden files/dirs (names starting with '.')
}
