package template

import (
	"path/filepath"

	"github.com/vietmpl/vie/analysis"
	"github.com/vietmpl/vie/value"
)

func (t Template) Analyze() (map[string]value.Type, []analysis.Diagnostic) {
	analyzer := analysis.NewAnalyzer()

	// TODO(skewb1k): process files concurrently.
	onFile := func(f *File, parent string) error {
		analyzer.Template(f.NameTemplate, "")
		if f.ContentTemplate != nil {
			analyzer.Template(f.ContentTemplate, filepath.Join(parent, f.Name))
		}
		return nil
	}
	onDir := func(d *Dir, parent string) error {
		analyzer.Template(d.NameTemplate, "")
		return nil
	}
	t.Walk(onDir, onFile)

	return analyzer.Results()
}
