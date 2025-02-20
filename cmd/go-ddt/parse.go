package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/xhd2015/data-driven-testing/pkgs/goresolve"
	"github.com/xhd2015/xgo/support/edit/goedit"
)

type FileEdit struct {
	astFile *AstFile
	vars    goresolve.Vars

	noWrite bool

	TargetFile *FileEdit

	edit          *goedit.Edit
	editHasUpdate bool
}

type TestCaseVar struct {
	VarName  string
	HasRef   bool
	TestCase *TestCase
}

type TestCase struct {
	Name      string
	Variants  []*Variant
	SubCases  []*TestCase
	HasAssert bool

	RefVarName string
	RefVar     *TestCaseVar
}

func (c *FileEdit) IsTestGo() bool {
	return strings.HasSuffix(c.astFile.File, "_test.go")
}

func (c *FileEdit) FileName() string {
	return c.astFile.File
}

func (c *FileEdit) GetEdit() *goedit.Edit {
	if c.edit != nil {
		return c.edit
	}
	c.edit = newEdit(c.astFile)
	return c.edit
}

func (c *FileEdit) GetFileEnd() token.Pos {
	return c.astFile.GetFileEnd()
}

func (c *FileEdit) EditAppend(code string) {
	if code == "" {
		return
	}
	end := c.GetFileEnd()
	c.GetEdit().Insert(end, code)
	c.editHasUpdate = true
}

func (c *FileEdit) EditHasUpdate() bool {
	return c.editHasUpdate
}

func (c *FileEdit) MarkEditUpdate() {
	c.editHasUpdate = true
}

type Variant struct {
	Name string
	Expr string

	ShortestName string
}

func getTestCaseVar(fset *token.FileSet, astFile *ast.File, code string, v *goresolve.Var) (*TestCaseVar, error) {
	if v == nil {
		return nil, nil
	}
	testCase, err := getTestCaseVarDef(fset, astFile, code, v.Def)
	if err != nil {
		return nil, err
	}
	return &TestCaseVar{
		VarName:  v.Name,
		HasRef:   v.HasRef,
		TestCase: testCase,
	}, nil
}

func getTestCaseVarDef(fset *token.FileSet, astFile *ast.File, code string, def *goresolve.Def) (*TestCase, error) {
	if def == nil {
		return nil, nil
	}
	var subCases []*TestCase
	for _, child := range def.Children {
		subCase, err := getTestCaseVarDef(fset, astFile, code, child)
		if err != nil {
			return nil, err
		}
		subCases = append(subCases, subCase)
	}
	var name string
	var variants []*Variant
	var hasAssert bool
	for _, field := range def.Fields {
		switch field.Name {
		case "Name":
			if basicLit, ok := field.Expr.(*ast.BasicLit); ok {
				var err error
				name, err = strconv.Unquote(basicLit.Value)
				if err != nil {
					return nil, err
				}
			}
		case "Assert":
			hasAssert = true
		case "Variants":
			variants = parseVariants(fset, field.Expr, code)

			names := make([]string, 0, len(variants))
			for _, v := range variants {
				names = append(names, v.Name)
			}
			shortestNames := shortestUncommonName(names)
			for i, v := range variants {
				v.ShortestName = shortestNames[i]
			}
		}
	}
	refVar, err := getTestCaseVar(fset, astFile, code, def.RefVar)
	if err != nil {
		return nil, err
	}
	return &TestCase{
		Name:      name,
		Variants:  variants,
		SubCases:  subCases,
		HasAssert: hasAssert,

		RefVarName: def.RefVarName,
		RefVar:     refVar,
	}, nil
}

func parseVariants(fset *token.FileSet, el ast.Expr, code string) []*Variant {
	if el == nil {
		return nil
	}
	var variants []*Variant
	valLit, ok := el.(*ast.CompositeLit)
	if ok {
		if _, ok := valLit.Type.(*ast.ArrayType); ok {
			for _, el := range valLit.Elts {
				variants = append(variants, &Variant{
					Name: exprAsName(el),
					Expr: exprToString(fset, el, code),
				})
			}
		}
	}

	return variants
}

func exprToString(fset *token.FileSet, el ast.Expr, code string) string {
	pos := fset.Position(el.Pos()).Offset
	end := fset.Position(el.End()).Offset

	return code[pos:end]
}

func exprAsName(el ast.Expr) string {
	// literal
	if el == nil {
		return ""
	}

	if basicLit, ok := el.(*ast.BasicLit); ok {
		if basicLit.Kind == token.STRING {
			s, _ := strconv.Unquote(basicLit.Value)
			return s
		}
		return basicLit.Value
	}

	if sel, ok := el.(*ast.SelectorExpr); ok {
		return sel.Sel.Name
	}

	if ident, ok := el.(*ast.Ident); ok {
		return ident.Name
	}

	return fmt.Sprintf("%v", el)
}

func shortestUncommonName(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	if len(names) == 1 {
		return []string{names[0]}
	}

	// Find longest common prefix
	prefix := names[0]
	for _, name := range names[1:] {
		for i := 0; i < len(prefix); i++ {
			if i >= len(name) || name[i] != prefix[i] {
				prefix = prefix[:i]
				break
			}
		}
		if prefix == "" {
			break
		}
	}

	// Find longest common suffix
	suffix := names[0]
	for _, name := range names[1:] {
		for i := 0; i < len(suffix); i++ {
			if i >= len(name) || name[len(name)-1-i] != suffix[len(suffix)-1-i] {
				suffix = suffix[len(suffix)-i:]
				break
			}
		}
		if suffix == "" {
			break
		}
	}

	// Trim both prefix and suffix from each name
	result := make([]string, len(names))
	for i, name := range names {
		trimmed := name
		if len(prefix) > 0 && strings.HasPrefix(trimmed, prefix) {
			trimmed = trimmed[len(prefix):]
		}
		if len(suffix) > 0 && strings.HasSuffix(trimmed, suffix) {
			trimmed = trimmed[:len(trimmed)-len(suffix)]
		}
		result[i] = trimmed
	}

	return result
}
