package format

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/etwicaksono/gotree/internal/config"
	"github.com/etwicaksono/gotree/internal/walker"
)

// RenderMarkdown produces the markdown document for the given tree and counts.
func RenderMarkdown(root string, cfg config.Config, tree *walker.Node, counts walker.Counts) string {
	var b bytes.Buffer

	// Header
	fmt.Fprintf(&b, "# Directory Tree: %s\n\n", root)
	fmt.Fprintf(&b, "- Generated at: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "- Max depth: %s\n", depthString(cfg.MaxDepth))
	fmt.Fprintf(&b, "- Follow symlinks: %t\n", cfg.FollowSymlinks)
	fmt.Fprintf(&b, "- Use .gitignore: %t\n", cfg.UseGitignore)
	fmt.Fprintf(&b, "- Exclude hidden: %t\n", cfg.ExcludeHidden)
	if len(cfg.Exclude) > 0 {
		fmt.Fprintf(&b, "- Exclude (any): %s\n", asCommaList(cfg.Exclude))
	}
	if len(cfg.ExcludeDir) > 0 {
		fmt.Fprintf(&b, "- Exclude dirs: %s\n", asCommaList(cfg.ExcludeDir))
	}
	if len(cfg.ExcludeFile) > 0 {
		fmt.Fprintf(&b, "- Exclude files: %s\n", asCommaList(cfg.ExcludeFile))
	}
	b.WriteString("\n")

	// Tree section
	b.WriteString("```text\n")
	renderTree(&b, tree, "", true)
	b.WriteString("```\n\n")
	if tree.IsDir && len(tree.Children) == 0 {
		b.WriteString("> No entries found under root with current filters.\n\n")
	} else if !tree.IsDir {
		b.WriteString("> Root is a file; no directory entries to list.\n\n")
	}

	// Counts
	fmt.Fprintf(&b, "- Directories: %d\n", counts.Dirs)
	fmt.Fprintf(&b, "- Files: %d\n", counts.Files)
	fmt.Fprintf(&b, "- Symlinks: %d\n", counts.Symlinks)

	return b.String()
}

func depthString(n int) string {
	if n <= 0 {
		return "unlimited"
	}
	return fmt.Sprintf("%d", n)
}

func asCommaList(items []string) string {
	cp := append([]string(nil), items...)
	sort.Slice(cp, func(i, j int) bool { return strings.ToLower(cp[i]) < strings.ToLower(cp[j]) })
	return strings.Join(cp, ", ")
}

// renderTree prints an ASCII tree starting at node.
// prefix is the current line prefix; isRoot indicates formatting of the first line.
func renderTree(b *bytes.Buffer, node *walker.Node, prefix string, isRoot bool) {
	// Print current node
	label := node.Name
	if node.IsDir {
		label += "/"
	}
	if node.IsSymlink && node.LinkTarget != "" {
		label += fmt.Sprintf(" -> %s", node.LinkTarget)
	}

	if isRoot {
		fmt.Fprintf(b, "%s\n", label)
	} else {
		fmt.Fprintf(b, "%s%s\n", prefix, label)
	}

	// Recurse for children
	n := len(node.Children)
	if n == 0 {
		return
	}

	for i, child := range node.Children {
		last := i == n-1
		conn := "├── "
		nextPrefix := prefix + "│   "
		if last {
			conn = "└── "
			nextPrefix = prefix + "    "
		}

		// Print child line header
		childLabel := child.Name
		if child.IsDir {
			childLabel += "/"
		}
		if child.IsSymlink && child.LinkTarget != "" {
			childLabel += fmt.Sprintf(" -> %s", child.LinkTarget)
		}
		fmt.Fprintf(b, "%s%s%s\n", prefix, conn, childLabel)

		// Recurse grandchildren with updated prefix
		if len(child.Children) > 0 {
			renderSubtree(b, child, nextPrefix)
		}
	}
}

// renderSubtree continues rendering for a non-root child using prefix rules.
func renderSubtree(b *bytes.Buffer, node *walker.Node, prefix string) {
	n := len(node.Children)
	for i, child := range node.Children {
		last := i == n-1
		conn := "├── "
		nextPrefix := prefix + "│   "
		if last {
			conn = "└── "
			nextPrefix = prefix + "    "
		}

		label := child.Name
		if child.IsDir {
			label += "/"
		}
		if child.IsSymlink && child.LinkTarget != "" {
			label += fmt.Sprintf(" -> %s", child.LinkTarget)
		}
		fmt.Fprintf(b, "%s%s%s\n", prefix, conn, label)

		if len(child.Children) > 0 {
			renderSubtree(b, child, nextPrefix)
		}
	}
}
