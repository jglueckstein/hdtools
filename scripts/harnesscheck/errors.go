package main

// A returned identifier err is the mechanical half of wrapped
// errors. Whether a panic is unreachable, and how os.IsNotExist is
// spelled, stay reviews. Tests are out of this check: helpers there
// return err on purpose.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func checkErrors(root string) ([]string, error) {
	var findings []string
	err := walkGo(root, false, func(path string) error {
		rel := relPath(root, path)
		if !under(rel, "cmd") && !under(rel, "internal") {
			return nil
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			findings = append(findings, fmt.Sprintf("%s: %v", rel, err))
			return nil
		}
		ast.Inspect(f, func(n ast.Node) bool {
			ret, ok := n.(*ast.ReturnStmt)
			if !ok {
				return true
			}
			for _, result := range ret.Results {
				id, ok := result.(*ast.Ident)
				if ok && id.Name == "err" {
					line := fset.Position(id.Pos()).Line
					findings = append(findings, fmt.Sprintf("%s:%d: returned err is not wrapped", rel, line))
				}
			}
			return true
		})
		return nil
	})
	return findings, err
}
