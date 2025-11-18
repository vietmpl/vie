package template_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vietmpl/vie/internal/template"
	"github.com/vietmpl/vie/value"
)

func TestTypes(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		context map[string]value.Value
	}

	cases := map[string]testCase{
		"filename-placeholder": {
			context: map[string]value.Value{
				"file": value.String("test"),
			},
		},
		"dirname-placeholder": {
			context: map[string]value.Value{
				"dir": value.String("test"),
			},
		},
	}

	for _, e := range entries {
		if !e.IsDir() || filepath.Ext(e.Name()) == ".golden" {
			continue
		}
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			inputPath := filepath.Join("testdata", name)

			tmpl, err := template.FromDir(inputPath)
			if err != nil {
				t.Fatal(err)
			}
			_, diagnostics := tmpl.Analyze()
			if len(diagnostics) > 0 {
				t.Fatalf("unexpected diagnostics %v", diagnostics)
			}
			println(name)
			context := cases[name].context
			files, err := tmpl.Render(context)
			if err != nil {
				t.Fatal(err)
			}

			expected := make(map[string][]byte)
			goldenPath := inputPath + ".golden"
			err = filepath.Walk(goldenPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				relPath, err := filepath.Rel(goldenPath, path)
				if err != nil {
					return err
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				expected[relPath] = data
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(expected, files); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
