package analysis_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/vietmpl/vie/analysis"
	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/parser"
	. "github.com/vietmpl/vie/value"
)

func TestTypes(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		Typemap     map[string]Type
		Diagnostics []Diagnostic
	}

	cases := map[string]testCase{
		"render": {
			Typemap: map[string]Type{
				"name": TypeString,
				"a":    TypeString,
				"b":    TypeString,
			},
			Diagnostics: []Diagnostic{
				WrongUsage{
					WantType: TypeString,
					GotType:  TypeBool,
					Pos_: ast.Pos{
						Line:      3,
						Character: 3,
					},
				},
			},
		},
		"if": {
			Typemap: map[string]Type{
				"a1": TypeBool,
				"a2": TypeBool,
				"b2": TypeBool,
				"c2": TypeBool,
			},
		},
		"binary": {
			Typemap: map[string]Type{
				"a": TypeBool,
				"b": TypeBool,
				"c": TypeBool,
				"d": TypeBool,
				"e": TypeBool,
				"f": TypeBool,
				"g": TypeBool,
			},
		},
		"parens": {
			Typemap: map[string]Type{
				"a": TypeString,
				"b": TypeString,
			},
		},
		"concatenate-bool-str": {
			Diagnostics: []Diagnostic{
				WrongUsage{
					WantType: TypeString,
					GotType:  TypeBool,
					Pos_: ast.Pos{
						Line:      0,
						Character: 8,
					},
				},
			},
		},
		"non-bool-if": {
			Diagnostics: []Diagnostic{
				WrongUsage{
					WantType: TypeBool,
					GotType:  TypeString,
					Pos_: ast.Pos{
						Line:      0,
						Character: 6,
					},
				},
			},
		},
		"non-bool-not": {
			Diagnostics: []Diagnostic{
				WrongUsage{
					WantType: TypeBool,
					GotType:  TypeString,
					Pos_: ast.Pos{
						Line:      0,
						Character: 7,
					},
				},
			},
		},
		"cross-var": {
			Diagnostics: []Diagnostic{
				CrossVarTyping{
					X: TypeVar("a"),
					Y: TypeVar("b"),
					Pos_: ast.Pos{
						Line:      0,
						Character: 6,
					},
				},
			},
		},
		"multiple-usages": {
			Typemap: map[string]Type{
				"a": TypeBool,
			},
			Diagnostics: []Diagnostic{
				WrongUsage{
					WantType: TypeString,
					GotType:  TypeBool,
					Pos_: ast.Pos{
						Line:      6,
						Character: 3,
					},
				},
			},
		},
		"equal-usages": {
			Typemap: map[string]Type{
				"a": TypeString,
			},
			Diagnostics: []Diagnostic{
				WrongUsage{
					WantType: TypeBool,
					GotType:  TypeString,
					Pos_: ast.Pos{
						Line:      0,
						Character: 6,
					},
				},
				WrongUsage{
					WantType: TypeBool,
					GotType:  TypeString,
					Pos_: ast.Pos{
						Line:      5,
						Character: 6,
					},
				},
			},
		},
		"call": {
			Typemap: map[string]Type{
				"a": TypeString,
			},
		},
		"pipe": {
			Typemap: map[string]Type{
				"a": TypeString,
			},
		},
		"incorrect-arg-count": {
			Diagnostics: []Diagnostic{
				IncorrectArgCount{
					FuncName: "@upper",
					Want:     1,
					Got:      2,
					Pos_: ast.Pos{
						Line:      0,
						Character: 3,
					},
				},
			},
		},
		"incorrect-arg-count-with-var": {
			Diagnostics: []Diagnostic{
				IncorrectArgCount{
					FuncName: "@upper",
					Want:     1,
					Got:      2,
					Pos_: ast.Pos{
						Line:      0,
						Character: 3,
					},
				},
			},
		},
		"func-not-found": {
			Diagnostics: []Diagnostic{
				BuiltinNotFound{
					Name: "@undefined_func",
					Msg:  "function @undefined_func is undefined",
					Pos_: ast.Pos{
						Line:      0,
						Character: 3,
					},
				},
			},
		},
		"call-undefined": {
			Diagnostics: []Diagnostic{
				BuiltinNotFound{
					Name: "@undefined_func",
					Msg:  "function @undefined_func is undefined",
					Pos_: ast.Pos{
						Line:      0,
						Character: 10,
					},
				},
			},
		},
		"wrong-use-call-undefined": {
			Diagnostics: []Diagnostic{
				BuiltinNotFound{
					Name: "@undefined_func",
					Msg:  "function @undefined_func is undefined",
					Pos_: ast.Pos{
						Line:      0,
						Character: 18,
					},
				},
			},
		},
	}

	for _, e := range entries {
		fileName := e.Name()
		name := fileName[:len(fileName)-len(".vie")]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			inputPath := filepath.Join("testdata", fileName)

			input, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatal(err)
			}

			f, err := parser.ParseBytes(input)
			if err != nil {
				t.Fatal(err)
			}

			want := cases[name]

			analyzer := NewAnalyzer()
			analyzer.File(f, "")
			typemap, diagnostics := analyzer.Results()
			got := testCase{
				Typemap:     typemap,
				Diagnostics: diagnostics,
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
