package main

import (
	"fmt"
	"strconv"
	"strings"
)

type TestCasePath []*TestCase

func (c *TestCase) getAllCases(parents TestCasePath) []TestCasePath {
	cases := make([]TestCasePath, 0, len(c.SubCases)+1)

	// copy
	casePath := make(TestCasePath, len(parents)+1)
	copy(casePath, parents)
	casePath[len(parents)] = c

	// only case with assert
	if c.HasAssert {
		cases = append(cases, casePath)
	}

	for _, subCase := range c.SubCases {
		subPrefixCases := subCase.getAllCases(casePath)
		cases = append(cases, subPrefixCases...)
	}
	return cases
}

func GetTestFuncName(path []string) string {
	names := make([]string, 0, len(path))
	for _, name := range path {
		if name == "" {
			name = "Unamed"
		}
		names = append(names, name)
	}
	return "Test" + JoinAsFuncName(names)
}

func FormatGoFunc(testFnName string, path []string, rootVar string) string {
	quoteNames := make([]string, 0, len(path))
	for _, name := range path {
		quoteNames = append(quoteNames, strconv.Quote(name))
	}

	quoteNameLit := strings.Join(quoteNames, ", ")
	return fmt.Sprintf(`func %s(t *testing.T) {
    %s.RunPath(t, []string{%s})
}`,
		testFnName,
		rootVar,
		quoteNameLit,
	)
}

func JoinAsFuncName(names []string) string {
	name := strings.Join(names, "-")
	name = strings.Title(name)
	return strings.ReplaceAll(name, "-", "_")
}

func genTestCases(varName string, casePaths []TestCasePath, verbose bool) []string {
	var genFuncs []string
	for _, casePath := range casePaths {
		names := make([]string, 0, len(casePath)+1)
		names = append(names, varName)
		for _, casePath := range casePath {
			names = append(names, casePath.Name)
		}
		testFnName := GetTestFuncName(names)
		if verbose {
			fmt.Printf("generate %s\n", testFnName)
		}
		fnCode := FormatGoFunc(testFnName, names[1:], varName)
		// hasUpdate = true
		// _ = fnCode
		// _ = insertPos
		genFuncs = append(genFuncs, PROLOG+"\n"+fnCode)
	}
	return genFuncs
}
