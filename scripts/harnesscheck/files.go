package main

// Walking is shared so each check sees the same set of Go files.
// Vendor and VCS directories are not source. Test files are included
// only when the check's rule names them.

import (
	"os"
	"path/filepath"
	"strings"
)

func walkGo(root string, includeTests bool, fn func(path string) error) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") {
			return nil
		}
		if !includeTests && strings.HasSuffix(name, "_test.go") {
			return nil
		}
		return fn(path)
	})
}

func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}

func under(rel, prefix string) bool {
	return rel == prefix || strings.HasPrefix(rel, prefix+string(filepath.Separator))
}
