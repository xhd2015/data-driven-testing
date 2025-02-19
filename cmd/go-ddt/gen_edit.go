package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func processGoFiles(dir string, verbose bool, singleFile string, dryRun bool) error {
	files, err := findGoFiles(dir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	fset := token.NewFileSet()
	fileEdits, err := parseFileEdits(fset, dir, files)
	if err != nil {
		return err
	}

	if err := parseAndResolveVars(fset, fileEdits); err != nil {
		return err
	}
	// delete all generated functions
	// in *_test.go
	for _, fileEdit := range fileEdits {
		astFile := fileEdit.astFile
		if singleFile != "" && fileEdit.FileName() != singleFile {
			if verbose {
				fmt.Printf("no write %s:\n", fileEdit.FileName())
			}
			fileEdit.noWrite = true
		}
		if !fileEdit.IsTestGo() {
			continue
		}
		if cleanGenEdit(fset, astFile.code, astFile.ast, fileEdit.GetEdit(fset)) {
			fileEdit.MarkEditUpdate()
		}
	}

	// correspond each file to its target file
	// i.e. if the file ends with _test.go, do nothing
	//      otherwise, create or find existing _test.go file for it
	generatedFiles, err := correspondTargetEditFiles(fset, dir, fileEdits)
	if err != nil {
		return err
	}

	// generate test cases to their target file
	for _, fileEdit := range fileEdits {
		generateTestCasesForFile(fset, fileEdit, verbose)
	}

	// write files
	allFiles := append(fileEdits, generatedFiles...)
	for _, fileEdit := range allFiles {
		fileName := fileEdit.FileName()
		if fileEdit.noWrite {
			continue
		}
		if !fileEdit.EditHasUpdate() {
			if verbose {
				fmt.Printf("no update %s\n", fileName)
			}
			continue
		}

		if verbose {
			fmt.Printf("updating %s\n", fileName)
		}
		if !dryRun {
			code := fileEdit.GetEdit(fset).String()
			err = os.WriteFile(filepath.Join(dir, fileName), []byte(code), 0755)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func findGoFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func parseFileEdits(fset *token.FileSet, dir string, files []string) ([]*FileEdit, error) {
	astFiles, err := ParseFiles(fset, dir, files)
	if err != nil {
		return nil, err
	}
	fileEdits := make([]*FileEdit, 0, len(astFiles))
	for _, astFile := range astFiles {
		fileEdits = append(fileEdits, &FileEdit{
			astFile: astFile,
		})
	}
	return fileEdits, nil
}

func parseAndResolveVars(fset *token.FileSet, fileEdits []*FileEdit) error {
	// parse test vars
	for _, fileEdit := range fileEdits {
		astFile := fileEdit.astFile
		astFileVars, err := pareFileVars(fset, astFile.ast, fileEdit.astFile.code)
		if err != nil {
			return err
		}
		fileEdit.vars = astFileVars
	}
	// resolve vars
	var allVars []*TestCaseVar
	for _, fileEdit := range fileEdits {
		allVars = append(allVars, fileEdit.vars...)
	}
	resolveVarRefs(allVars)
	return nil
}

func generateTestCasesForFile(fset *token.FileSet, fileEdit *FileEdit, verbose bool) {
	targetFile := fileEdit.TargetFile
	if targetFile == nil {
		// itself
		targetFile = fileEdit
	}
	var needImportTesting bool
	for _, vr := range fileEdit.vars {
		if vr.HasRef {
			continue
		}
		varGenFuncs := genTestCases(vr.VarName, vr.TestCase.getAllCases(nil), verbose)
		for i, genFunc := range varGenFuncs {
			if genFunc == "" {
				continue
			}
			var suffix string
			if i < len(varGenFuncs)-1 {
				suffix = "\n\n"
			}
			targetFile.EditAppend(fset, "\n"+genFunc+suffix)
			needImportTesting = true
		}
	}
	if needImportTesting {
		if verbose {
			fmt.Printf("importing testing for %s\n", targetFile.FileName())
		}
		importPkg(fset, targetFile, "testing")
	}
}

func correspondTargetEditFiles(fset *token.FileSet, dir string, fileEdits []*FileEdit) (generatedFiles []*FileEdit, err error) {
	fileEditMapping := make(map[string]*FileEdit, len(fileEdits))
	for _, fileEdit := range fileEdits {
		fileEditMapping[fileEdit.astFile.file] = fileEdit
	}

	// generate placeholder files for each var
	for _, fileEdit := range fileEdits {
		var targetFile *FileEdit
		if len(fileEdit.vars) == 0 || fileEdit.IsTestGo() {
			continue
		}
		fileName := fileEdit.FileName()
		testGoFile := strings.TrimSuffix(fileName, ".go") + "_test.go"
		targetFile = fileEditMapping[testGoFile]
		if targetFile == nil {
			pkgName := fileEdit.astFile.ast.Name.Name
			testCode := fmt.Sprintf("package %s\n", pkgName)
			testGoAst, err := ParseCode(fset, dir, testGoFile, testCode)
			if err != nil {
				return nil, err
			}
			targetFile = &FileEdit{
				astFile: testGoAst,
			}
			generatedFiles = append(generatedFiles, targetFile)
			fileEditMapping[testGoFile] = targetFile
		}
		// inherit nowrite
		targetFile.noWrite = fileEdit.noWrite
		fileEdit.TargetFile = targetFile
	}
	return generatedFiles, nil
}

func importPkg(fset *token.FileSet, fileEdit *FileEdit, pkg string) {
	pkgQuote := strconv.Quote(pkg)
	astFile := fileEdit.astFile.ast

	// Check if testing is already imported
	for _, imp := range astFile.Imports {
		if imp.Path != nil && imp.Path.Value == pkgQuote {
			return // already imported
		}
	}

	// Find the position to insert import and determine the format
	var insertPos token.Pos
	var importStmt string

	// Look for existing import declarations
	var lastImportDecl *ast.GenDecl
	for _, decl := range astFile.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			lastImportDecl = genDecl
		}
	}

	if lastImportDecl != nil {
		// If there's an existing import declaration
		if lastImportDecl.Lparen == token.NoPos {
			// Single import without parentheses - convert to parenthesized form
			insertPos = lastImportDecl.End()
			importStmt = fmt.Sprintf("import (\n\t%s\n\t%s\n)", lastImportDecl.Specs[0].(*ast.ImportSpec).Path.Value, pkgQuote)
			// Delete the original import
			edit := fileEdit.GetEdit(token.NewFileSet())
			edit.Delete(lastImportDecl.Pos(), lastImportDecl.End())
		} else {
			// Already has parentheses
			insertPos = lastImportDecl.Rparen
			// Find the indentation of the last import
			var lastImportPos token.Pos
			if n := len(lastImportDecl.Specs); n > 0 {
				lastImportPos = lastImportDecl.Specs[n-1].Pos()
			}
			if lastImportPos != token.NoPos {
				// Get the indentation from the last import
				lastImportOffset := fset.Position(lastImportPos).Offset
				lineStart := lastImportOffset
				for lineStart > 0 && fileEdit.astFile.code[lineStart-1] != '\n' {
					lineStart--
				}
				indent := fileEdit.astFile.code[lineStart:lastImportOffset]
				importStmt = fmt.Sprintf("%s%s\n", indent, pkgQuote)
			} else {
				importStmt = fmt.Sprintf("\t%s\n", pkgQuote)
			}
		}
	} else {
		// No existing imports
		insertPos = astFile.Name.End()
		importStmt = fmt.Sprintf("\n\nimport %s", pkgQuote)
	}

	// Add the import
	edit := fileEdit.GetEdit(fset)
	edit.Insert(insertPos, importStmt)
	fileEdit.MarkEditUpdate()
}
