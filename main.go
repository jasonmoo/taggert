package main

import (
	"flag"
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
		Name   string `json:"name"bson:"nome,omitempty"`
		Height string `json:"hieght"bson:"height"`
	}
)

var (
	anagram        = flag.Bool("anagram", false, "check for anagrams (commonly misspellings)")
	un_camel_case  = flag.Bool("un_camel_case", false, "match field names to tag names un-camelcased.  Ex: WalterWhite -> walter_white")
	all_tags_match = flag.Bool("all_tags_match", false, "ensure all tags match")
	levenshtein    = flag.Int("levenshtein", 0, "report if levenshtein distance > n")

	quotedValueRegexp = regexp.MustCompile(`([^:]+):"([^",]+)[^"]*"`)
	unCamelCaseRegexp = regexp.MustCompile(`([A-Z]+)`)
)

func init() {
	flag.Parse()
	if flag.NFlag() == 0 || flag.NArg() == 0 {
		fmt.Println("taggert usage:")
		fmt.Println("./taggert [-flags] source.go")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {

	w := &tabwriter.Writer{}
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	defer w.Flush()

	fset := token.NewFileSet()

	for _, file := range flag.Args() {

		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			fmt.Println(err)
			return
		}

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
						matches := quotedValueRegexp.FindAllStringSubmatch(field.Tag.Value[1:], -1)
						for _, match := range matches {
							if len(match) != 3 {
								continue
							}
							if msg := Test(field.Names[0].String(), match[2]); msg != "" {
								fmt.Fprintf(
									w,
									"%s:%d\tField name %s\t%q\t%s:%q\n",
									file,
									fset.File(field.Pos()).Line(field.Pos()),
									msg,
									field.Names[0].String(),
									match[1],
									match[2],
								)
							}
						}
						if *all_tags_match && len(matches) > 1 {
							for i := 1; i < len(matches); i++ {
								if msg := Test(matches[0][1], matches[i][1]); msg != "" {
									fmt.Fprintf(
										w,
										"%s:%d\tTag name %s\t%s:%q\t%s:%q\n",
										file,
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
					}
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
	if *un_camel_case {
		a, b = ccToUS(a), ccToUS(b)
	}
	if *anagram && IsAnagram(a, b) {
		return "anagram"
	}
	if a != b {
		return "mismatch"
	}
	return ""
}

func IsAnagram(a, b string) bool {
	a, b = strings.ToLower(a), strings.ToLower(b)
	if len(a) != len(b) || a == b {
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
