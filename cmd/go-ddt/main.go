package main

import (
	"fmt"
	"os"
	"strings"
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
	case "help", "--help":
		fmt.Println(strings.TrimSpace(help))
		return nil
	case "gen":
		return handleGen(args[1:])
	case "view":
		return handleView(args[1:])
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
	// example env:
	//   GOFILE=my_program_file.go
	//   GOLINE=21
	//   GOPACKAGE=my_program
	goEnvFile := os.Getenv("GOFILE")
	goEnvLine := os.Getenv("GOLINE")
	goEnvPackage := os.Getenv("GOPACKAGE")

	// log env
	// fmt.Printf("DEBUG: GOFILE: %s\n", goEnvFile)
	// fmt.Printf("DEBUG: GOLINE: %s\n", goEnvLine)
	// fmt.Printf("DEBUG: GOPACKAGE: %s\n", goEnvPackage)

	// single go generate mode, change only one file
	var singleFile string
	if goEnvFile != "" && goEnvLine != "" && goEnvPackage != "" {
		singleFile = goEnvFile
		if verbose {
			fmt.Printf("go generate on: %s\n", singleFile)
		}
	}
	return processGoFiles(dir, verbose, singleFile, dryRun)
}
