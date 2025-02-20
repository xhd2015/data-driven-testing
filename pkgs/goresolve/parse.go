package goresolve

import (
	"fmt"
	"go/ast"
	"go/token"
)

func ParseVars(fset *token.FileSet, astFile *ast.File, code string, childrenKey string) (Vars, error) {
	var vars Vars

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
			var parseDefOrRef func(el ast.Expr) (*Def, error)
			parseDefOrRef = func(el ast.Expr) (*Def, error) {
				switch el := el.(type) {
				case *ast.CompositeLit:
					if _, ok := el.Type.(*ast.ArrayType); ok {

						// the variable itself is a slice
						var children []*Def
						for _, child := range el.Elts {
							childDef, err := parseDefOrRef(child)
							if err != nil {
								return nil, fmt.Errorf("%w", err)
							}
							children = append(children, childDef)
						}
						return &Def{
							Children: children,
						}, nil
					}
					return parseVarCmpositeLit(fset, el, code, childrenKey)
				case *ast.Ident:
					return &Def{
						RefVarName: el.Name,
					}, nil
				default:
					// nothing to do with
					return nil, nil
				}
			}
			def, err := parseDefOrRef(el)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", varName, err)
			}
			vars = append(vars, &Var{
				Name: varName,
				Def:  def,
			})
		}
	}
	return vars, nil
}

func parseVarCmpositeLit(fset *token.FileSet, compLit *ast.CompositeLit, code string, childrenKey string) (*Def, error) {
	if compLit == nil {
		return nil, nil
	}
	var fields []*Field
	var children []*Def
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
		case childrenKey:
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
							subCase, err := parseVarCmpositeLit(fset, p, code, childrenKey)
							if err != nil {
								return nil, err
							}
							children = append(children, subCase)
						case *ast.Ident:
							children = append(children, &Def{
								RefVarName: p.Name,
							})
						default:
							return nil, fmt.Errorf("unrecognized: %T %v", p, p)
						}
					}
				}
			}
		default:
			fields = append(fields, &Field{
				Name: keyName,
				Expr: kv.Value,
			})
		}
	}
	return &Def{
		Fields:   fields,
		Children: children,
	}, nil
}
