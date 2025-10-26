package walker

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/etwicaksono/gotree/internal/config"
	"github.com/etwicaksono/gotree/internal/filter"
)

type Node struct {
	Name       string
	Rel        string
	IsDir      bool
	IsSymlink  bool
	LinkTarget string
	Children   []*Node
}

type Counts struct {
	Dirs     int
	Files    int
	Symlinks int
}

// Walk traverses the filesystem starting at rootPath using cfg and filt,
// producing a tree of Nodes and aggregate counts.
// Depth semantics: depth 0 = root. If cfg.MaxDepth > 0, traversal will include
// entries up to depth N and not descend into depth N+1.
func Walk(rootPath string, cfg config.Config, filt *filter.Filter) (*Node, Counts, error) {
	var counts Counts

	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, counts, err
	}
	st, err := os.Lstat(absRoot)
	if err != nil {
		return nil, counts, err
	}
	isSymlink := st.Mode()&os.ModeSymlink != 0

	rootRel := "."
	rootNode := &Node{
		Name:      filepath.Base(absRoot),
		Rel:       rootRel,
		IsDir:     st.IsDir(),
		IsSymlink: isSymlink,
	}
	if isSymlink {
		counts.Symlinks++
		if tgt, err := os.Readlink(absRoot); err == nil {
			rootNode.LinkTarget = tgt
		}
	}
	if rootNode.IsDir {
		counts.Dirs++
	} else {
		counts.Files++
	}

	visited := newVisited()
	err = buildTree(absRoot, rootNode, 0, cfg, filt, visited, &counts)
	if err != nil {
		return nil, counts, err
	}

	return rootNode, counts, nil
}

func buildTree(absPath string, node *Node, depth int, cfg config.Config, filt *filter.Filter, visited *visitedSet, counts *Counts) error {
	// If not a directory, nothing to do
	if !node.IsDir {
		return nil
	}

	// Respect max depth
	if cfg.MaxDepth > 0 && depth >= cfg.MaxDepth {
		return nil
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		// If directory cannot be read (permissions, etc.), just stop here
		return nil
	}

	// Sort entries by name for stable output
	sort.Slice(entries, func(i, j int) bool { return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name()) })

	for _, de := range entries {
		name := de.Name()
		childAbs := filepath.Join(absPath, name)
		rel, _ := filepath.Rel(filepath.Clean(absPath), childAbs) // relative to parent for now
		if parentRel := node.Rel; parentRel != "." {
			rel = filepath.ToSlash(filepath.Join(parentRel, name))
		} else {
			rel = filepath.ToSlash(name)
		}

		// Use Lstat to detect symlinks
		fi, err := os.Lstat(childAbs)
		if err != nil {
			continue
		}
		isSymlink := fi.Mode()&os.ModeSymlink != 0

		// Determine if this entry is a directory without following symlinks by default
		isDir := de.IsDir()
		linkTarget := ""
		if isSymlink {
			if tgt, err := os.Readlink(childAbs); err == nil {
				linkTarget = tgt
			}
		}

		// Filtering
		e := filter.Entry{
			Path:      childAbs,
			Rel:       rel,
			Name:      name,
			IsDir:     isDir,
			IsSymlink: isSymlink,
		}
		if filt != nil && filt.ShouldExclude(e) {
			// Prune excluded entry (and its subtree if dir)
			continue
		}

		child := &Node{
			Name:       name,
			Rel:        rel,
			IsDir:      isDir,
			IsSymlink:  isSymlink,
			LinkTarget: linkTarget,
		}

		// Update counts
		if isSymlink {
			counts.Symlinks++
		}
		if isDir {
			counts.Dirs++
		} else {
			counts.Files++
		}

		// Append child
		node.Children = append(node.Children, child)

		// Traverse child if directory
		nextAbs := childAbs
		nextDepth := depth + 1

		// Follow symlink directories only when configured
		if isSymlink {
			if cfg.FollowSymlinks {
				// If symlink, attempt to resolve and if target is a dir, traverse
				targetInfo, statErr := os.Stat(childAbs)
				if statErr != nil {
					continue
				}
				if targetInfo.IsDir() {
					real, canErr := canonicalPath(childAbs)
					if canErr != nil {
						continue
					}
					if visited.Seen(real) {
						continue
					}
					visited.Add(real)
					// Treat as directory for traversal
					if cfg.MaxDepth == 0 || nextDepth <= cfg.MaxDepth {
						// ensure flag for child is directory during traversal
						child.IsDir = true
						if err := buildTree(childAbs, child, nextDepth, cfg, filt, visited, counts); err != nil {
							return err
						}
					}
				}
			}
			// If not following symlinks, do not traverse
			continue
		}

		// Normal directory traversal
		if isDir {
			real, canErr := canonicalPath(nextAbs)
			if canErr == nil {
				if visited.Seen(real) {
					continue
				}
				visited.Add(real)
			}
			if err := buildTree(nextAbs, child, nextDepth, cfg, filt, visited, counts); err != nil {
				return err
			}
		}
	}

	return nil
}

type visitedSet struct {
	mu  sync.Mutex
	set map[string]struct{}
}

func newVisited() *visitedSet {
	return &visitedSet{set: make(map[string]struct{})}
}

func (v *visitedSet) Add(path string) {
	v.mu.Lock()
	v.set[path] = struct{}{}
	v.mu.Unlock()
}

func (v *visitedSet) Seen(path string) bool {
	v.mu.Lock()
	_, ok := v.set[path]
	v.mu.Unlock()
	return ok
}

// canonicalPath tries to produce a canonical absolute path for cycle detection.
func canonicalPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// If cannot eval symlink, fall back to abs path
		// Return error only if abs also fails (handled above)
		return abs, nil
	}
	if real == "" {
		return "", errors.New("empty canonical path")
	}
	return real, nil
}
