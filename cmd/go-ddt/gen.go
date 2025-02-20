package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
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

func (c TestCasePath) GetEffectiveVariants() []*Variant {
	n := len(c)
	for i := n - 1; i >= 0; i-- {
		if len(c[i].Variants) > 1 {
			return c[i].Variants
		}
	}
	return nil
}

func GetTestFuncName(path []string, variant *Variant) string {
	identNames := make([]string, 0, len(path))
	for _, name := range path {
		if name == "" {
			name = "Unamed"
		}
		identName := nameToIdentifierSuffix(compactName(name))
		identName = capitalize(identName)
		identNames = append(identNames, identName)
	}
	if variant != nil && variant.ShortestName != "" {
		identName := nameToIdentifierSuffix(compactName(variant.ShortestName))
		identName = capitalize(identName)
		identNames = append(identNames, identName)
	}
	return "Test" + JoinAsFuncName(identNames)
}

func compactName(name string) string {
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")
	return name
}

func nameToIdentifierSuffix(name string) string {
	// Replace special characters with similar Unicode symbols that are valid Go identifiers
	name = strings.ReplaceAll(name, ">", "ᐳ") // U+1433 CANADIAN SYLLABICS PO
	name = strings.ReplaceAll(name, "<", "ᐸ") // U+1438 CANADIAN SYLLABICS PA
	name = strings.ReplaceAll(name, "=", "ᗕ") // U+15D5 CANADIAN SYLLABICS CARRIER CHI
	name = strings.ReplaceAll(name, "%", "ᵖ") // U+1D56 MODIFIER LETTER SMALL P
	name = strings.ReplaceAll(name, "+", "ᐩ") // U+1429 CANADIAN SYLLABICS FINAL PLUS
	name = strings.ReplaceAll(name, "-", "ˉ") // U+02C9 MODIFIER LETTER MACRON
	name = strings.ReplaceAll(name, "*", "ˣ") // U+02E3 MODIFIER LETTER SMALL X
	name = strings.ReplaceAll(name, "/", "ᐟ") // U+141F CANADIAN SYLLABICS FINAL ACUTE

	var result strings.Builder
	for _, ch := range name {
		// Subsequent characters can be letters, digits, or underscore
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

func capitalize(name string) string {
	if len(name) == 0 {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func FormatGoFunc(testFnName string, path []string, rootVar string, variant *Variant) string {
	quoteNames := make([]string, 0, len(path))
	for _, name := range path {
		quoteNames = append(quoteNames, strconv.Quote(name))
	}

	fnName := "RunPath"
	extraArgs := ""
	if variant != nil {
		fnName = "RunPathVariant"
		extraArgs = fmt.Sprintf(", %s", variant.Expr)
	}

	quoteNameLit := strings.Join(quoteNames, ", ")
	return fmt.Sprintf(`func %s(t *testing.T) {
    %s.%s(t, []string{%s}%s)
}`,
		testFnName,
		rootVar,
		fnName,
		quoteNameLit,
		extraArgs,
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
		effectiveVariants := casePath.GetEffectiveVariants()
		if len(effectiveVariants) > 0 {
			// generate variants
			for _, variant := range effectiveVariants {
				_, fnCode := generateTestFunction(varName, casePath, variant, verbose)
				genFuncs = append(genFuncs, fnCode)
			}
		} else {
			_, fnCode := generateTestFunction(varName, casePath, nil, verbose)
			genFuncs = append(genFuncs, fnCode)
		}
	}
	return genFuncs
}

func generateTestFunction(varName string, casePath TestCasePath, variant *Variant, verbose bool) (string, string) {
	names := make([]string, 0, len(casePath)+1)
	names = append(names, varName)
	for _, casePath := range casePath {
		names = append(names, casePath.Name)
	}
	testFnName := GetTestFuncName(names, variant)
	if verbose {
		fmt.Printf("generate %s\n", testFnName)
	}
	fnCode := FormatGoFunc(testFnName, names[1:], varName, variant)
	return testFnName, PROLOG + "\n" + fnCode
}
