package main

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const help = `
go-ddt generate data driven testing cases.

Usage: go-ddt <cmd> [OPTIONS] <ARGS>

Commands:
  gen 

Options:
    --dir DIR    directory
    --dry-run    dry run
 -v,--verbose    show verbose info
    --help       show help message

Examples:
  $ go-ddt gen
  $ go-ddt gen ./...
`

const VERSION = "0.0.1"

func main() {
	err := handle(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
func handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires cmd")
	}
	cmd := args[0]
	switch cmd {
	case "version":
		fmt.Println(VERSION)
		return nil
	case "help":
		fmt.Println(strings.TrimSpace(help))
		return nil
	case "gen":
		return handleGen(args[1:])
	default:
		return fmt.Errorf("unrecognized command: %s", cmd)
	}
}

func handleGen(args []string) error {
	var dir string

	var verbose bool
	var dryRun bool
	var remainArgs []string
	n := len(args)
	for i := 0; i < n; i++ {
		if args[i] == "--dir" {
			if i+1 >= n {
				return fmt.Errorf("%v requires arg", args[i])
			}
			dir = args[i+1]
			i++
			continue
		}
		if args[i] == "--help" {
			fmt.Println(strings.TrimSpace(help))
			return nil
		}
		if args[i] == "--verbose" || args[i] == "-v" {
			verbose = true
			continue
		}
		if args[i] == "--dry-run" {
			dryRun = true
			continue
		}
		if args[i] == "--" {
			remainArgs = append(remainArgs, args[i+1:]...)
			break
		}
		if strings.HasPrefix(args[i], "-") {
			return fmt.Errorf("unrecognized flag: %v", args[i])
		}
		remainArgs = append(remainArgs, args[i])
	}

	if dir == "" {
		dir = "./"
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			files = append(files, entry.Name())
		}
	}
	// fmt.Println(files)

	if len(files) == 0 {
		return nil
	}

	fset := token.NewFileSet()
	astFiles, err := ParseFiles(fset, dir, files)
	if err != nil {
		return err
	}
	// collect all vars
	fileVars := make([]*FileVar, 0, len(astFiles))
	for _, astFile := range astFiles {
		astFileVars, err := pareFileVars(astFile.ast)
		if err != nil {
			return err
		}
		fileVars = append(fileVars, &FileVar{
			astFile: astFile,
			vars:    astFileVars,
		})
	}

	// resolve vars
	allVars := make([]*TestCaseVar, 0, len(fileVars))
	for _, fileVar := range fileVars {
		allVars = append(allVars, fileVar.vars...)
	}
	resolveVarRefs(allVars)

	// delete all generated functions
	fileEdits := make([]*FileVarEdit, 0, len(fileVars))
	for _, fileVar := range fileVars {
		edit := fileVar.astFile.newEdit(fset)
		newCode, hasUpdate := cleanGen(fset, fileVar.astFile.code, fileVar.astFile.ast, edit)

		if hasUpdate {
			// trim suffix space
			newCode = strings.TrimRightFunc(newCode, unicode.IsSpace)
		}
		fileEdits = append(fileEdits, &FileVarEdit{
			fileVar:   fileVar,
			hasUpdate: hasUpdate,
			code:      newCode,
		})
	}

	// regenerate functions
	for _, fileEdit := range fileEdits {
		var fileGenFuncs []string
		for _, vr := range fileEdit.fileVar.vars {
			if vr.HasRef {
				continue
			}
			varGenFuncs := genTestCases(vr.VarName, vr.TestCase.getAllCases(nil), verbose)
			fileGenFuncs = append(fileGenFuncs, varGenFuncs...)
		}
		if len(fileGenFuncs) == 0 {
			continue
		}
		fileEdit.hasUpdate = true
		fileEdit.code += "\n" + strings.Join(fileGenFuncs, "\n\n")
	}

	// write file
	for _, fileEdit := range fileEdits {
		if !fileEdit.hasUpdate {
			continue
		}
		file := fileEdit.fileVar.astFile.file
		if verbose {
			fmt.Printf("updating %s\n", file)
		}

		if !dryRun {
			err = os.WriteFile(filepath.Join(dir, file), []byte(fileEdit.code), 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
