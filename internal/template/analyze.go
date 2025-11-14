package template

import (
	"github.com/vietmpl/vie/analysis"
	"github.com/vietmpl/vie/value"
)

func (t Template) Analyze() (map[string]value.Type, []analysis.Diagnostic) {
	analyzer := analysis.NewAnalyzer()

	// TODO(skewb1k): process files concurrently.
	onFile := func(f *File, parent string) error {
		analyzer.File(f.NameAST)
		if f.ContentAST != nil {
			analyzer.File(f.ContentAST)
		}
		return nil
	}
	onDir := func(d *Dir, parent string) error {
		analyzer.File(d.NameAST)
		return nil
	}
	t.Walk(onDir, onFile)

	return analyzer.Results()
}
