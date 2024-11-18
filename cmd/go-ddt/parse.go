package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
)

type FileVarEdit struct {
	fileVar   *FileVar
	hasUpdate bool
	code      string
}

type FileVar struct {
	astFile *astFile
	vars    []*TestCaseVar
}

type TestCaseVar struct {
	VarName  string
	HasRef   bool
	TestCase *TestCase
}

type TestCase struct {
	Name      string
	SubCases  []*TestCase
	HasAssert bool

	RefVarName string
	RefVar     *TestCaseVar
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

func pareFileVars(astFile *ast.File) ([]*TestCaseVar, error) {
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
				testCase, err := parseCmpositeLit(el)
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

func parseCmpositeLit(compLit *ast.CompositeLit) (*TestCase, error) {
	if compLit == nil {
		return nil, nil
	}
	var name string
	var subCases []*TestCase
	var hasAssert bool
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
							subCase, err := parseCmpositeLit(p)
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
	}, nil
}
