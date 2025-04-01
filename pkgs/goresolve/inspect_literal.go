package goresolve

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// GetCompositeProps extracts AST nodes by their names from an already parsed file
// keys is a list of identifiers to extract (function names, variable names, etc.)
// returns a map of identifier name -> ast.Node
func GetCompositeProps(lit *ast.CompositeLit, keys []string) map[string]ast.Node {
	if lit == nil {
		return nil
	}
	result := make(map[string]ast.Node, len(keys))
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		idt, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		var found bool
		for _, key := range keys {
			if idt.Name == key {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		result[idt.Name] = kv.Value
	}
	return result
}

func FuncLitToNamed(fset *token.FileSet, node ast.Node, content string, funcName string) string {
	nodeContent := NodeToString(fset, node, content)
	if nodeContent == "" {
		return ""
	}

	return fmt.Sprintf("func %s%s", funcName, strings.TrimPrefix(nodeContent, "func"))
}

// NodeToString converts an AST node back to source code
func NodeToString(fset *token.FileSet, node ast.Node, content string) string {
	if node == nil {
		return ""
	}

	startPos := fset.Position(node.Pos())
	endPos := fset.Position(node.End())

	if startPos.Offset >= 0 && endPos.Offset > startPos.Offset && endPos.Offset <= len(content) {
		code := content[startPos.Offset:endPos.Offset]
		lines := strings.Split(code, "\n")

		if len(lines) > 1 {
			lastLine := lines[len(lines)-1]
			// count spaces before }
			spaces := countPrefixSpaces(lastLine)
			if spaces > 0 {
				// for each line, remove the spaces up to the number of spaces in the last line
				for i, line := range lines {
					lines[i] = trimPrefixSpaceAtMost(line, spaces)
				}
			}
			return strings.Join(lines, "\n")
		}

		return code
	}

	return ""
}

// trimPrefixSpaceAtMost removes at most 'spaces' number of leading whitespace characters
func trimPrefixSpaceAtMost(line string, spaces int) string {
	i := 0
	for _, char := range line {
		if char == ' ' || char == '\t' {
			i++
			if i >= spaces {
				break
			}
		} else {
			break
		}
	}
	return line[i:]
}

// countPrefixSpaces counts the number of leading whitespace characters
func countPrefixSpaces(line string) int {
	i := 0
	for _, char := range line {
		if char == ' ' || char == '\t' {
			i++
		} else {
			break
		}
	}
	return i
}
