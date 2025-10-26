package output

import (
	"os"
	"path/filepath"
)

// WriteOutput writes content to the given path, ensuring parent directories exist.
// Overwrites the target file if it already exists.
// UTF-8 is assumed by caller when providing content string.
func WriteOutput(path string, content string) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
