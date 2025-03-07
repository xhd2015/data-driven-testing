package main

import (
	"github.com/xhd2015/data-driven-testing/pkgs/goast"

	"github.com/xhd2015/xgo/support/edit/goedit"
)

const PROLOG = "// generated by go-ddt, DO NOT EDIT."

type AstFile = goast.AstFile

func newEdit(c *AstFile) *goedit.Edit {
	return goedit.New(c.Fset, c.Code)
}
