package main

// Layout checks where files live, not which type is the package's
// main export. Production code is cmd/hdtools and internal/.
// scripts/ holds these check tools and is not a third application
// tree. A test is *_test.go next to that code.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func checkLayout(root string) ([]string, error) {
	var findings []string
	mainPath := filepath.Join(root, "cmd", "hdtools", "main.go")
	if _, err := os.Stat(mainPath); err != nil {
		findings = append(findings, "cmd/hdtools/main.go: missing")
	}
	err := walkGo(root, true, func(path string) error {
		rel := relPath(root, path)
		if under(rel, "scripts") {
			return nil
		}
		base := filepath.Base(path)
		isTest := strings.HasSuffix(base, "_test.go")
		if !under(rel, filepath.Join("cmd", "hdtools")) && !under(rel, "internal") {
			kind := "production file"
			if isTest {
				kind = "test file"
			}
			findings = append(findings, fmt.Sprintf("%s: %s is outside cmd/hdtools and internal", rel, kind))
		}
		return nil
	})
	return findings, err
}
