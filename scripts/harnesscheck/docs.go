package main

// Missing doc comments are the mechanical half of exported API
// documentation. Whether the first sentence says what, rather than
// how, is a review. A const or type group with one comment documents
// that group, which is the Go idiom.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func checkDocs(root string) ([]string, error) {
	var findings []string
	err := walkGo(root, false, func(path string) error {
		rel := relPath(root, path)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: %v", rel, err))
			return nil
		}
		for _, decl := range f.Decls {
			switch n := decl.(type) {
			case *ast.FuncDecl:
				if n.Name.IsExported() && n.Doc == nil {
					line := fset.Position(n.Name.Pos()).Line
					findings = append(findings, fmt.Sprintf("%s:%d: exported %s has no doc comment", rel, line, n.Name.Name))
				}
			case *ast.GenDecl:
				if n.Tok != token.TYPE && n.Tok != token.CONST {
					continue
				}
				if n.Doc != nil {
					continue
				}
				for _, spec := range n.Specs {
					name, doc := specNameDoc(spec)
					if name != nil && name.IsExported() && doc == nil {
						line := fset.Position(name.Pos()).Line
						findings = append(findings, fmt.Sprintf("%s:%d: exported %s has no doc comment", rel, line, name.Name))
					}
				}
			}
		}
		return nil
	})
	return findings, err
}

func specNameDoc(spec ast.Spec) (*ast.Ident, *ast.CommentGroup) {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		return s.Name, s.Doc
	case *ast.ValueSpec:
		if len(s.Names) == 0 {
			return nil, s.Doc
		}
		return s.Names[0], s.Doc
	default:
		return nil, nil
	}
}
