package goresolve

import "go/ast"

type Vars []*Var

type Var struct {
	Name   string
	HasRef bool
	Def    *Def
}

type Def struct {
	Fields   []*Field
	Children []*Def

	RefVarName string
	RefVar     *Var
}

type Field struct {
	Name string
	Expr ast.Expr
}
