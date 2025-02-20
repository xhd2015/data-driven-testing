package goast

import (
	"go/ast"
	"go/token"
)

type AstFile struct {
	Fset *token.FileSet
	File string // short name: something.go
	Code string
	Ast  *ast.File
}

func (c *AstFile) GetFileStart() token.Pos {
	return getFileStart(c.Fset, len(c.Code), c.Ast)
}

func (c *AstFile) GetFileEnd() token.Pos {
	return getFileEnd(c.Fset, len(c.Code), c.Ast)
}
