package t_tree_static

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

// ParseFile parses a Go file into an AST file
func ParseFile(fset *token.FileSet, file string) (*ast.File, []byte, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}
	astFile, err := ParseCode(fset, string(bytes))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse AST: %w", err)
	}
	return astFile, bytes, nil
}

// ParseCode parses Go code content into an AST file
func ParseCode(fset *token.FileSet, content string) (*ast.File, error) {
	// Try parsing as a Go file
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go code: %v", err)
	}

	return f, nil
}
