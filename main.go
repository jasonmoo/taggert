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
	"text/tabwriter"
	"unicode/utf8"
)

type (
	Dog struct {
		Name   string `json:"name"bson:"nome"`
		Height string `json:"hieght"bson:"height"`
	}
)

var (
	quotedValueRegexp = regexp.MustCompile(`([^:]+):"([^"]+?)(?:,omitempty)?"`)
	unCamelCaseRegexp = regexp.MustCompile(`([A-Z]+)`)
)

func main() {

	fset := token.NewFileSet() // positions are relative to fset
	w := &tabwriter.Writer{}
	w.Init(os.Stdout, 0, 12, 0, '\t', 0)
	defer w.Flush()

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

				if field.Tag != nil {
					us_name := ccToUS(field.Names[0].String())
					matches := quotedValueRegexp.FindAllStringSubmatch(field.Tag.Value[1:], -1)
					for _, match := range matches {
						if len(match) != 3 {
							continue
						}
						if msg := Test(us_name, match[2]); msg != "" {
							fmt.Fprintf(
								w,
								"%s:%d\tField name %s\t\t%s\t%s:%q\n",
								os.Args[1],
								fset.File(field.Pos()).Line(field.Pos()),
								msg,
								field.Names[0].String(),
								match[1],
								match[2],
							)
						}
					}
					if len(matches) > 1 {
						for i := 1; i < len(matches); i++ {
							if msg := Test(matches[0][1], matches[i][1]); msg != "" {
								fmt.Fprintf(
									w,
									"%s:%d\tTag name incongruency\t%s\t%s:%q\t%s:%q\n",
									os.Args[1],
									fset.File(field.Pos()).Line(field.Pos()),
									msg,
									matches[0][1],
									matches[0][2],
									matches[i][1],
									matches[i][2],
								)
							}
						}

					}

					// if len(matches) == 2 {
					// 	if msg := Test(matches[0][2], matches[1][2], field.Names[0].String()); msg != "" {
					// 		fmt.Printf(
					// 			"%s:%d  %s (%s): %s:%q : %s:%q\n",
					// 			os.Args[1],
					// 			fset.File(field.Pos()).Line(field.Pos()),
					// 			msg,
					// 			field.Names[0].String(),
					// 			// uc_name,
					// 			// match[1],
					// 			matches[0][1],
					// 			matches[0][2],
					// 			matches[1][1],
					// 			matches[1][2],
					// 		)
					// 	}
					// }
				}
			}
		}
	}
}

func ccToUS(s string) string {
	return strings.TrimPrefix(unCamelCaseRegexp.ReplaceAllStringFunc(s, func(s string) string {
		return "_" + strings.ToLower(s)
	}), "_")
}

func Test(a, b string) string {
	a, b = ccToUS(a), ccToUS(b)
	if IsAnagram(a, b) {
		return "anagram"
	}
	if a != b {
		return "mismatch"
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
