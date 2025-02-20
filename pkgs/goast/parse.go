package goast

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func ParseFiles(fset *token.FileSet, dir string, files []string) ([]*AstFile, error) {
	astFiles := make([]*AstFile, 0, len(files))
	for _, file := range files {
		astFile, err := ParseFile(fset, dir, file)
		if err != nil {
			return nil, err
		}
		astFiles = append(astFiles, astFile)
	}
	return astFiles, nil
}

func ParseFile(fset *token.FileSet, dir string, file string) (*AstFile, error) {
	srcFile := filepath.Join(dir, file)

	code, err := os.ReadFile(srcFile)
	if err != nil {
		return nil, err
	}
	parsedAST, err := parser.ParseFile(fset, srcFile, code, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return &AstFile{
		Fset: fset,
		File: file,
		Code: string(code),
		Ast:  parsedAST,
	}, nil
}

func ParseCode(fset *token.FileSet, dir string, file string, code string) (*AstFile, error) {
	parsedAST, err := parser.ParseFile(fset, filepath.Join(dir, file), code, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return &AstFile{
		Fset: fset,
		File: file,
		Code: string(code),
		Ast:  parsedAST,
	}, nil
}
