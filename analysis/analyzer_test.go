package analysis_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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
		types       map[string]Type
		diagnostics []Diagnostic
	}

	cases := map[string]testCase{
		"render": {
			types: map[string]Type{
				"name": TypeString,
				"a":    TypeString,
				"b":    TypeString,
			},
			diagnostics: []Diagnostic{
				&WrongUsage{
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
			types: map[string]Type{
				"a1": TypeBool,
				"a2": TypeBool,
				"b2": TypeBool,
				"c2": TypeBool,
			},
		},
		"binary": {
			types: map[string]Type{
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
			types: map[string]Type{
				"a": TypeString,
				"b": TypeString,
			},
		},
		"concatenate-bool-str": {
			types: map[string]Type{},
			diagnostics: []Diagnostic{
				&WrongUsage{
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
			types: map[string]Type{},
			diagnostics: []Diagnostic{
				&WrongUsage{
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
			types: map[string]Type{},
			diagnostics: []Diagnostic{
				&WrongUsage{
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
			types: map[string]Type{},
			diagnostics: []Diagnostic{
				&CrossVarTyping{
					X: VarType("a"),
					Y: VarType("b"),
					Pos_: ast.Pos{
						Line:      0,
						Character: 6,
					},
				},
			},
		},
		"multiple-usages": {
			types: map[string]Type{
				"a": TypeBool,
			},
			diagnostics: []Diagnostic{
				&WrongUsage{
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
			types: map[string]Type{
				"a": TypeString,
			},
			diagnostics: []Diagnostic{
				&WrongUsage{
					WantType: TypeBool,
					GotType:  TypeString,
					Pos_: ast.Pos{
						Line:      0,
						Character: 6,
					},
				},
				&WrongUsage{
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
			types: map[string]Type{
				"a": TypeString,
			},
		},
		"incorrect-arg-count": {
			types: map[string]Type{},
			diagnostics: []Diagnostic{
				&IncorrectArgCount{
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
			types: map[string]Type{},
			diagnostics: []Diagnostic{
				&IncorrectArgCount{
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
			types: map[string]Type{},
			diagnostics: []Diagnostic{
				&BuiltinNotFound{
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
			types: map[string]Type{},
			diagnostics: []Diagnostic{
				&BuiltinNotFound{
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
			types: map[string]Type{},
			diagnostics: []Diagnostic{
				&BuiltinNotFound{
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

			sf := parser.Source(input)

			actualTypes, actualDiagnostics := File(sf)

			assert.Equal(t, cases[name], testCase{
				types:       actualTypes,
				diagnostics: actualDiagnostics,
			})
		})
	}
}
