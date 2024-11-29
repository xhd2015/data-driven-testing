//go:build go1.20
// +build go1.20

package main

import (
	"go/ast"
	"go/token"
)

func getFileEnd(fset *token.FileSet, fileSize int, astFile *ast.File) token.Pos {
	return astFile.FileEnd
}
