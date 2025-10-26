# gotree

A Go CLI that generates a markdown-formatted directory and file tree.

## Overview
This tool traverses a root path, applies glob-based excludes, optionally respects `.gitignore`, and writes a Markdown file containing an ASCII tree plus metadata.

## Features
- Glob excludes for files and directories (doublestar v4)
- Optional `.gitignore` support via `--gitignore`
- Depth limiting with `--max-depth`
- Symlink listing by default; optional traversal via `--follow-symlinks` with cycle protection
- Hidden items included by default; opt-out via `--exclude-hidden`
- Deterministic output ordering
- Markdown output with header, flags summary, tree, and counts

## Requirements
- Go 1.25+ (module targets 1.25.1)

## Install
If using from source:
```sh
git clone https://github.com/etwicaksono/gotree
cd gotree
go run . --help
```

When you just need the binary, you can install:
```sh
go install github.com/etwicaksono/gotree@latest
```

## Build
Build a local binary:
```sh
go build -o gotree .
./gotree --help
```

## Usage
Basic command:
```sh
gotree [flags]
```

## Flags
```text
--root string           Root path to traverse (default: .)
--output string         Markdown output file path (default: ./TREE.md)
--exclude string        Glob pattern applied to names (repeatable)
--exclude-dir string    Glob pattern applied only to directories (repeatable)
--exclude-file string   Glob pattern applied only to files (repeatable)
--max-depth int         Maximum depth to traverse (0 = unlimited)
--gitignore             Apply rules from .gitignore at root
--follow-symlinks       Traverse directory symlinks
--exclude-hidden        Exclude hidden files and directories
--version               Print version and exit
--help                  Print help and exit
```

## Filtering semantics
- Precedence per entry: hidden (if enabled) → `.gitignore` (if enabled) → type-specific globs → global globs.
- First match excludes the entry, directories are pruned from traversal.
- Glob engine: `bmatcuk/doublestar/v4` supports `**`, character classes, and path-aware matching.
- Patterns match against both the base name and the slash-normalized relative path.

## Depth and symlinks
- Depth 0 = root only; each nested level increments by 1.
- With `--max-depth N`, entries up to depth N are included; traversal does not descend into N+1.
- Symlinks are listed as `name -> target`; traversal only when `--follow-symlinks`.
- Cycles are prevented via a canonical visited set.

## Examples
- Generate with defaults:
```sh
gotree --root . --output ./TREE.md
```

- Exclude `node_modules` and `*.log` files:
```sh
gotree --exclude node_modules --exclude "*.log"
```

- Apply `.gitignore` and limit depth to 3:
```sh
gotree --gitignore --max-depth 3
```

- Exclude only files by pattern:
```sh
gotree --exclude-file "*.min.js" --exclude-file "*.map"
```

- Exclude specific directories and follow symlinks:
```sh
gotree --exclude-dir vendor --exclude-dir .cache --follow-symlinks
```

- Exclude hidden items:
```sh
gotree --exclude-hidden
```

## Output format
Markdown file contains:
- Title and metadata header (timestamp, active flags)
- ASCII tree within a fenced code block
- Counts for directories, files, symlinks

### Sample
```markdown
# Directory Tree: .

- Generated at: 2025-10-26T20:00:40Z
- Max depth: unlimited
- Follow symlinks: false
- Use .gitignore: false
- Exclude hidden: false

```text
project/
├── cmd/
│   └── root.go
└── main.go
```

- Directories: 2
- Files: 2
- Symlinks: 0
```

## Project structure
- CLI entry: cmd/root.go
- Main: main.go
- Pipeline:
  - Config: internal/config/config.go
  - Filter: internal/filter/filter.go
  - Walker: internal/walker/walker.go
  - Formatter: internal/format/markdown.go
  - Writer: internal/output/writer.go
  - Version: internal/version/version.go
