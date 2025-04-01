package goresolve

import (
	"go/ast"
	"go/token"
	"strconv"
)

type FindLiteralOptions struct {
	StartLine int
	EndLine   int
}

// FindMatchingLiteral finds `"<key>":"<value>"` pair in a composite literal
// within a given range `startLine`~`endLine`
// returns the node if found, otherwise returns nil
// the `key` is usually an `ID` or `Key` that identifies the node
func FindMatchingLiteral(fset *token.FileSet, astFile *ast.File, key string, value string, options FindLiteralOptions) *ast.CompositeLit {
	startLine := options.StartLine
	endLine := options.EndLine

	var foundNode *ast.CompositeLit

	// Traverse the AST to find the node matching the criteria
	ast.Inspect(astFile, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		// Check if the node is within the specified line range
		pos := fset.Position(n.Pos())
		end := fset.Position(n.End())

		// Skip if not in the line range
		if end.Line < startLine || (endLine > 0 && pos.Line > endLine) {
			return true
		}

		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		var foundKV *ast.KeyValueExpr
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == key {
				foundKV = kv
				break
			}
		}
		if foundKV == nil || foundKV.Value == nil {
			return true
		}
		litValue, ok := foundKV.Value.(*ast.BasicLit)
		if !ok {
			return true
		}
		if litValue.Kind != token.STRING {
			return true
		}
		unquoted, parseErr := strconv.Unquote(litValue.Value)
		if parseErr != nil {
			return true
		}
		if unquoted != value {
			return true
		}

		// found the kv
		foundNode = lit
		return false
	})

	return foundNode
}
