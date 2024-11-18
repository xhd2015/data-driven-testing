package main

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/xhd2015/xgo/support/edit/goedit"
)

func cleanGen(fset *token.FileSet, code string, astFile *ast.File, edit *goedit.Edit) (string, bool) {
	progLines := make(map[int]*ast.Comment)
	for _, cmt := range astFile.Comments {
		for _, cm := range cmt.List {
			if strings.HasPrefix(cm.Text, PROLOG) {
				line := fset.Position(cm.Pos()).Line
				progLines[line] = cm
			}
		}
	}

	var hasUpdate bool
	n := len(astFile.Decls)

	fileEnd := astFile.FileEnd
	for i := 0; i < n; i++ {
		fnDecl, ok := astFile.Decls[i].(*ast.FuncDecl)
		if !ok {
			continue
		}
		fnLine := fset.Position(fnDecl.Pos()).Line
		cm := progLines[fnLine-1]
		if cm == nil {
			continue
		}
		hasUpdate = true

		// delete all subsequent empty spaces
		edit.Delete(cm.Pos(), fnDecl.End())
		for p := fnDecl.End(); p < fileEnd; p++ {
			idx := fset.Position(p).Offset
			if idx >= len(code) || !isSpace(code[idx]) {
				break
			}
			edit.Delete(p, p+1)
		}
	}

	if !hasUpdate {
		return code, false
	}

	return edit.String(), true
}

func isSpace(c byte) bool {
	return c == '\n' || c == ' ' || c == '\t'
}
