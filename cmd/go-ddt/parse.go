package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/xhd2015/xgo/support/edit/goedit"
)

type FileEdit struct {
	astFile *astFile
	vars    []*TestCaseVar

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
	return strings.HasSuffix(c.astFile.file, "_test.go")
}

func (c *FileEdit) FileName() string {
	return c.astFile.file
}

func (c *FileEdit) GetEdit(fset *token.FileSet) *goedit.Edit {
	if c.edit != nil {
		return c.edit
	}
	c.edit = c.astFile.newEdit(fset)
	return c.edit
}

func (c *FileEdit) GetFileEnd(fset *token.FileSet) token.Pos {
	return getFileEnd(fset, len(c.astFile.code), c.astFile.ast)
}

func (c *FileEdit) EditAppend(fset *token.FileSet, code string) {
	if code == "" {
		return
	}
	end := c.GetFileEnd(fset)
	c.GetEdit(fset).Insert(end, code)
	c.editHasUpdate = true
}

func (c *FileEdit) EditHasUpdate() bool {
	return c.editHasUpdate
}

func (c *FileEdit) MarkEditUpdate() {
	c.editHasUpdate = true
}

func resolveVarRefs(vars []*TestCaseVar) error {
	mappingByNames := make(map[string]*TestCaseVar, len(vars))
	for _, v := range vars {
		mappingByNames[v.VarName] = v
	}

	var traverse func(tc *TestCase) error

	traverse = func(tc *TestCase) error {
		refVarName := tc.RefVarName
		if refVarName != "" {
			refVar := mappingByNames[refVarName]
			if refVar == nil {
				return fmt.Errorf("%s not found", refVarName)
			}
			refVar.HasRef = true
			tc.RefVar = refVar
		}
		for _, subCase := range tc.SubCases {
			err := traverse(subCase)
			if err != nil {
				return err
			}
		}
		return nil

	}
	for _, v := range vars {
		err := traverse(v.TestCase)
		if err != nil {
			return fmt.Errorf("%s: %w", v.VarName, err)
		}
	}
	return nil
}

func pareFileVars(fset *token.FileSet, astFile *ast.File, code string) ([]*TestCaseVar, error) {
	var testCaseVars []*TestCaseVar

	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			valSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			names := valSpec.Names
			values := valSpec.Values
			if len(names) != 1 {
				continue
			}
			if len(values) != 1 {
				continue
			}

			varName := names[0].Name
			if varName == "" {
				continue
			}

			var el ast.Expr = values[0]
			if unaryExpr, ok := el.(*ast.UnaryExpr); ok && unaryExpr.Op == token.AND {
				// deref &
				el = unaryExpr.X
			}

			switch el := el.(type) {
			case *ast.CompositeLit:
				testCase, err := parseCmpositeLit(fset, el, code)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", varName, err)
				}
				testCaseVars = append(testCaseVars, &TestCaseVar{
					VarName:  varName,
					TestCase: testCase,
				})
			case *ast.Ident:
				testCaseVars = append(testCaseVars, &TestCaseVar{
					VarName: varName,
					TestCase: &TestCase{
						RefVarName: el.Name,
					},
				})
			default:
				// nothing to do with
			}
		}
	}
	return testCaseVars, nil
}

func parseCmpositeLit(fset *token.FileSet, compLit *ast.CompositeLit, code string) (*TestCase, error) {
	if compLit == nil {
		return nil, nil
	}
	var name string
	var subCases []*TestCase
	var hasAssert bool
	var variants []*Variant
	for _, el := range compLit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		idt, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		keyName := idt.Name
		switch keyName {
		case "Name":
			basicLit, ok := kv.Value.(*ast.BasicLit)
			if ok {
				var err error
				name, err = strconv.Unquote(basicLit.Value)
				if err != nil {
					return nil, err
				}
			}
		case "Assert":
			hasAssert = true
		case "Variants":
			variants = parseVariants(fset, kv.Value, code)

			names := make([]string, 0, len(variants))
			for _, v := range variants {
				names = append(names, v.Name)
			}
			shortestNames := shortestUncommonName(names)
			for i, v := range variants {
				v.ShortestName = shortestNames[i]
			}
		case "SubCases":
			valLit, ok := kv.Value.(*ast.CompositeLit)
			if ok {
				if _, ok := valLit.Type.(*ast.ArrayType); ok {
					for _, el := range valLit.Elts {
						p := el
						if unary, ok := p.(*ast.UnaryExpr); ok && unary.Op == token.AND {
							p = unary.X
						}
						switch p := p.(type) {
						case *ast.CompositeLit:
							subCase, err := parseCmpositeLit(fset, p, code)
							if err != nil {
								return nil, err
							}
							subCases = append(subCases, subCase)
						case *ast.Ident:
							subCases = append(subCases, &TestCase{
								RefVarName: p.Name,
							})
						default:
							return nil, fmt.Errorf("unrecognized: %T %v", p, p)
						}
					}
				}
			}
		}
	}
	return &TestCase{
		Name:      name,
		HasAssert: hasAssert,
		SubCases:  subCases,
		Variants:  variants,
	}, nil
}

type Variant struct {
	Name string
	Expr string

	ShortestName string
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
