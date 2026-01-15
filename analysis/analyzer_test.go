package analysis_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/vietmpl/vie/analysis"
	"github.com/vietmpl/vie/parse"
	"github.com/vietmpl/vie/value"
)

func TestTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input       string
		typemap     map[string]value.Type
		diagnostics []analysis.Diagnostic
	}{
		{
			input: "{{ name }}",
			typemap: map[string]value.Type{
				"name": value.TypeString,
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()

			f, err := parse.Source([]byte(testCase.input))
			if err != nil {
				t.Fatal(err)
			}

			analyzer := analysis.NewAnalyzer()
			analyzer.Template(f, "")
			typemap, diagnostics := analyzer.Results()

			if !maps.Equal(testCase.typemap, typemap) {
				t.Errorf("expected %v, got %v", testCase.typemap, typemap)
			}

			if !slices.Equal(testCase.diagnostics, diagnostics) {
				t.Errorf("expected %v, got %v", testCase.diagnostics, diagnostics)
			}
		})
	}
}
