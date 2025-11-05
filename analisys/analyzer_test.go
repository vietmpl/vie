package analisys

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/parser"
	. "github.com/vietmpl/vie/value"
)

func TestTypes(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	type tc struct {
		types       map[string]Type
		diagnostics []Diagnostic
	}

	cases := map[string]tc{
		"render": {
			types: map[string]Type{
				"name": TypeString,
				"a":    TypeString,
				"b":    TypeString,
			},
			diagnostics: []Diagnostic{
				&WrongUsage{
					ExpectedType: TypeString,
					GotType:      TypeBool,
					_Pos: ast.Pos{
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
	}

	for _, e := range entries {
		fileName := e.Name()
		name := fileName[:len(fileName)-len(".vie")]
		t.Run(name, func(t *testing.T) {
			inputPath := filepath.Join("testdata", fileName)

			input, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatal(err)
			}

			sf := parser.ParseFile(input)

			gotTypes, gotDiagnostics := Source(sf)

			got := tc{
				types:       gotTypes,
				diagnostics: gotDiagnostics,
			}

			assert.Equal(t, got, cases[name])
		})
	}
}
