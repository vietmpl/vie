package analisys_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vietmpl/vie/analisys"
	"github.com/vietmpl/vie/parser"
)

func TestTypes(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		types       map[string]analisys.Type
		diagnostics []analisys.Diagnostic
	}{
		"render": {
			types: map[string]analisys.Type{
				"name": analisys.TypeString,
				"a":    analisys.TypeString,
				"b":    analisys.TypeString,
			},
		},
		"if": {
			types: map[string]analisys.Type{
				"a1": analisys.TypeBool,
				"a2": analisys.TypeBool,
				"b2": analisys.TypeBool,
				"c2": analisys.TypeBool,
			},
		},
		"binary": {
			types: map[string]analisys.Type{
				"a": analisys.TypeBool,
				"b": analisys.TypeBool,
				"c": analisys.TypeBool,
				"d": analisys.TypeBool,
				"e": analisys.TypeBool,
				"f": analisys.TypeBool,
				"g": analisys.TypeBool,
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

			sourceFile, err := parser.ParseFile(input)
			if err != nil {
				t.Fatal(err)
			}

			gotTypes, gotDiagnostics := analisys.Source(sourceFile)

			want := cases[name]

			if !reflect.DeepEqual(gotTypes, want.types) {
				t.Errorf("mismatch for %s\n--- got\n%v\n--- want\n%v", name, gotTypes, want.types)
			}
			if !reflect.DeepEqual(gotDiagnostics, want.diagnostics) {
				t.Errorf("mismatch for %s\n--- got\n%v\n--- want\n%v", name, gotDiagnostics, want.diagnostics)
			}
		})
	}
}
