package t_tree_static

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/xhd2015/data-driven-testing/pkgs/goresolve"
)

// GetNodeProps handles the request to extract just the Assert function from a test case
func GetNodeProps(fset *token.FileSet, astFile *ast.File, nodeID string, props []string) (map[string]ast.Node, error) {
	if len(props) == 0 {
		return nil, fmt.Errorf("require props")
	}
	lit, err := FindNodeInFile(fset, astFile, nodeID)
	if err != nil {
		return nil, err
	}

	return goresolve.GetCompositeProps(lit, props), nil
}

// GetNodePropsAsFuncs handles the request to extract just the Assert function from a test case
func GetNodePropsAsFuncs(fset *token.FileSet, astFile *ast.File, code string, nodeID string, keys []string) (map[string]string, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("requires nodeID")
	}

	props, err := GetNodeProps(fset, astFile, nodeID, keys)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if props[key] == nil {
			continue
		}
		code := goresolve.FuncLitToNamed(fset, props[key], code, key)
		if code == "" {
			continue
		}
		result[key] = code
	}
	return result, nil
}
