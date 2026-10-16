package main

// An opening comment is the mechanical half of the literate
// preamble. In this tree the narrative sits either immediately
// before the package clause or between that clause and the imports.
// Whether the words answer why, the decisions, and what the file
// does not do is a review.

import (
	"fmt"
	"go/parser"
	"go/token"
)

func checkPreamble(root string) ([]string, error) {
	var findings []string
	err := walkGo(root, false, func(path string) error {
		rel := relPath(root, path)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: %v", rel, err))
			return nil
		}
		var first token.Pos
		if len(f.Imports) > 0 {
			first = f.Imports[0].Pos()
		} else if len(f.Decls) > 0 {
			first = f.Decls[0].Pos()
		}
		for _, group := range f.Comments {
			if group.End() < f.Package {
				return nil
			}
			if first.IsValid() && group.Pos() > f.Package && group.End() < first {
				return nil
			}
		}
		findings = append(findings, rel+": no opening comment")
		return nil
	})
	return findings, err
}
