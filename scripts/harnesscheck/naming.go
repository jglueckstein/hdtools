package main

// File names and package names are the mechanical half of Idiomatic
// Go naming. MixedCaps, initialisms, and a Get prefix are reviews.
// An external test package may be the directory name plus _test.
// That suffix is the Go idiom, not a second word.

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"unicode"
)

func checkNaming(root string) ([]string, error) {
	var findings []string
	err := walkGo(root, true, func(path string) error {
		rel := relPath(root, path)
		base := filepath.Base(path)
		if !strings.HasSuffix(base, "_test.go") && strings.Contains(base, "_") {
			findings = append(findings, rel+": file name contains '_'")
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly)
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: %v", rel, err))
			return nil
		}
		name := f.Name.Name
		stem := name
		if strings.HasSuffix(base, "_test.go") {
			stem = strings.TrimSuffix(name, "_test")
		}
		if strings.Contains(stem, "_") || hasUpper(stem) {
			line := fset.Position(f.Name.Pos()).Line
			findings = append(findings, fmt.Sprintf("%s:%d: package %s", rel, line, name))
		}
		return nil
	})
	return findings, err
}

func hasUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}
