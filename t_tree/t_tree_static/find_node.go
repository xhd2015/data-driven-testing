package t_tree_static

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/xhd2015/data-driven-testing/pkgs/goresolve"
)

// FindNodeInFile searches for a node definition in a file
func FindNodeInFile(fset *token.FileSet, astFile *ast.File, nodeID string) (*ast.CompositeLit, error) {
	literal := goresolve.FindMatchingLiteral(fset, astFile, "ID", nodeID, goresolve.FindLiteralOptions{})
	if literal == nil {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	return literal, nil
}
