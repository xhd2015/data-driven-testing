//go:build go1.20
// +build go1.20

package goast

import (
	"go/ast"
	"go/token"
)

func getFileStart(fset *token.FileSet, fileSize int, astFile *ast.File) token.Pos {
	return astFile.FileStart
}

func getFileEnd(fset *token.FileSet, fileSize int, astFile *ast.File) token.Pos {
	return astFile.FileEnd
}
