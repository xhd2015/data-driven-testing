//go:build !go1.20
// +build !go1.20

package goast

import (
	"go/ast"
	"go/token"
)

func getFileStart(fset *token.FileSet, fileSize int, astFile *ast.File) token.Pos {
	pos := fset.Position(astFile.Package)
	fileStart := astFile.Package - token.Pos(pos.Offset)
	return fileStart
}

func getFileEnd(fset *token.FileSet, fileSize int, astFile *ast.File) token.Pos {
	pos := fset.Position(astFile.Package)
	fileStart := astFile.Package - token.Pos(pos.Offset)
	return fileStart + token.Pos(fileSize)
}
