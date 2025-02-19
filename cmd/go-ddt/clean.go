package main

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/xhd2015/xgo/support/edit/goedit"
)

func cleanGenEdit(fset *token.FileSet, code string, astFile *ast.File, edit *goedit.Edit) bool {
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

	fileEnd := getFileEnd(fset, len(code), astFile)
	var lastDelEndOffset int // avoid overlap
	for i := 0; i < n; i++ {
		fnDecl, ok := astFile.Decls[i].(*ast.FuncDecl)
		if !ok {
			continue
		}
		fnLine := fset.Position(fnDecl.Pos()).Line
		commentLine := progLines[fnLine-1]
		if commentLine == nil {
			continue
		}
		hasUpdate = true

		commentStartPos := commentLine.Pos()
		delStartPos := commentStartPos
		// delete one precendent empty space
		commentStartIdx := fset.Position(commentStartPos).Offset
		if commentStartIdx > 0 && commentStartIdx-1 >= lastDelEndOffset && isNewLine(code[commentStartIdx-1]) {
			delStartPos = commentStartPos - 1
		}

		// delete all subsequent empty spaces
		endPos := fnDecl.End()
		endPosOffset := fset.Position(endPos).Offset
		edit.Delete(delStartPos, endPos)
		lastDelEndOffset = endPosOffset
		for p := endPos; p < fileEnd; p++ {
			offset := fset.Position(p).Offset
			if offset >= len(code) || !isSpace(code[offset]) {
				break
			}
			edit.Delete(p, p+1)
			lastDelEndOffset = offset + 1
		}
	}
	return hasUpdate
}

func isSpace(c byte) bool {
	return c == '\n' || c == ' ' || c == '\t'
}

func isNewLine(c byte) bool {
	return c == '\n'
}
