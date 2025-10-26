package filter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	ignore "github.com/sabhiram/go-gitignore"

	"github.com/etwicaksono/gotree/internal/config"
)

// Entry describes a filesystem entry for filtering decisions.
type Entry struct {
	// Absolute path on disk
	Path string
	// Path relative to root, slash-normalized
	Rel string
	// Base name of the entry
	Name string
	// Entry is a directory
	IsDir bool
	// Entry is a symlink
	IsSymlink bool
}

// Filter encapsulates exclude logic and optional .gitignore matcher.
type Filter struct {
	cfg        config.Config
	root       string
	gitIgnore  *ignore.GitIgnore
	hasGitFile bool
}

// New constructs a Filter from config, optionally compiling .gitignore at root.
func New(root string, cfg config.Config) (*Filter, error) {
	f := &Filter{
		cfg:  cfg,
		root: root,
	}
	if cfg.UseGitignore {
		gitPath := filepath.Join(root, ".gitignore")
		if st, err := os.Stat(gitPath); err == nil && !st.IsDir() {
			gi, err := ignore.CompileIgnoreFile(gitPath)
			if err != nil {
				return nil, err
			}
			f.gitIgnore = gi
			f.hasGitFile = true
		}
	}
	return f, nil
}

// ShouldExclude returns true if the entry should be excluded and, for directories, pruned.
func (f *Filter) ShouldExclude(e Entry) bool {
	// 1) Hidden exclusion (names starting with '.')
	if f.cfg.ExcludeHidden && strings.HasPrefix(e.Name, ".") {
		return true
	}

	// 2) .gitignore rules (only when enabled and file exists)
	if f.cfg.UseGitignore && f.hasGitFile && f.gitIgnore != nil {
		// gitignore matches using slash paths relative to the git root
		if f.gitIgnore.MatchesPath(e.Rel) {
			return true
		}
	}

	// 3) Type-specific excludes
	if e.IsDir && matchAny(f.cfg.ExcludeDir, e.Name, e.Rel) {
		return true
	}
	if !e.IsDir && matchAny(f.cfg.ExcludeFile, e.Name, e.Rel) {
		return true
	}

	// 4) Global excludes
	if matchAny(f.cfg.Exclude, e.Name, e.Rel) {
		return true
	}

	return false
}

func matchAny(patterns []string, name string, rel string) bool {
	rel = toSlash(rel)
	for _, p := range patterns {
		if p == "" {
			continue
		}
		// Match against both base name and slash-normalized relative path for flexibility
		if pathMatch(p, name) || pathMatch(p, rel) {
			return true
		}
	}
	return false
}

func toSlash(p string) string {
	if p == "" {
		return p
	}
	return filepath.ToSlash(p)
}

func pathMatch(pattern, path string) bool {
	// doublestar treats '/' as path separator consistently across platforms
	ok, _ := doublestar.PathMatch(pattern, path)
	return ok
}
