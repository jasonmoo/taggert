package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

type (
	Dog struct {
		Name string `json:"name"bson:"nome"`
	}
)

var (
	quotedValueRegexp = regexp.MustCompile(`([^:]+):"([^"]+?)(?:,omitempty)?"`)
	unCamelCaseRegexp = regexp.MustCompile(`([A-Z]+)`)
)

func Test(a, b, name string) string {
	if IsAnagram(a, b) {
		return "anagram found"
	}
	if strings.Trim(a, "_") != strings.Trim(b, "_") {
		return "mismatch"
	}
	us_name := strings.TrimPrefix(unCamelCaseRegexp.ReplaceAllStringFunc(name, func(s string) string {
		return "_" + strings.ToLower(s)
	}), "_")
	if us_name != a || us_name != b {
		return "unexpected name"
	}
	return ""
}

func IsAnagram(a, b string) bool {
	if a == b || len(a) != len(b) {
		return false
	}
	ai := make([]int, utf8.RuneCountInString(a))
	bi := make([]int, utf8.RuneCountInString(b))
	for i, r := range a {
		ai[i] = int(r)
	}
	for i, r := range b {
		bi[i] = int(r)
	}
	sort.Ints(ai)
	sort.Ints(bi)
	for i := 0; i < len(ai); i++ {
		if ai[i] != bi[i] {
			return false
		}
	}
	return true
}

func main() {

	fset := token.NewFileSet() // positions are relative to fset

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, os.Args[1], nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the imports from the file's AST.
	for _, s := range f.Decls {

		node, ok := s.(*ast.GenDecl)
		if !ok || node.Tok != token.TYPE {
			continue
		}

		for _, spec := range node.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			stct, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}

			for _, field := range stct.Fields.List {

				if len(field.Names) != 1 {
					continue
				}

				// fmt.Printf("%s : %s\n", field.Names[0].String(), uc_name)
				if field.Tag != nil {
					matches := quotedValueRegexp.FindAllStringSubmatch(field.Tag.Value, -1)
					if len(matches) == 2 {
						if msg := Test(matches[0][2], matches[1][2], field.Names[0].String()); msg != "" {
							fmt.Printf(
								"%s:%d  %s (%s): %s:%q : %s:%q\n",
								os.Args[1],
								fset.File(field.Pos()).Line(field.Pos()),
								msg,
								field.Names[0].String(),
								// uc_name,
								// match[1],
								matches[0][1],
								matches[0][2],
								matches[1][1],
								matches[1][2],
							)
						}
					}
				}
			}
		}
	}
}
