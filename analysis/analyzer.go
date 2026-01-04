package analysis

import (
	"sync"

	"github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/value"
)

type Analyzer struct {
	mu          sync.RWMutex
	usages      map[string][]Usage
	diagnostics []Diagnostic
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		mu:          sync.RWMutex{},
		usages:      make(map[string][]Usage),
		diagnostics: nil,
	}
}

// Template analyzes a single file and records usages/diagnostics.
// TODO(skewb1k): support context.
// TODO(skewb1k): remove 'path' argument.
func (a *Analyzer) Template(template *ast.Template, path string) {
	c := internalContext{
		path: path,
	}
	a.checkBlocks(c, template.Blocks)
}

// Results completes type inference and returns results.
// Should be called once after all files are analyzed.
//
// During analysis, every occurrence of an identifier is recorded in
// [analyzer.usages] along with the type expected in that context. After
// traversal of the AST is complete, this function aggregates all collected
// usages to infer each variable's most likely type and emit diagnostics for
// all other wrong usages.
func (a *Analyzer) Results() (map[string]value.Type, []Diagnostic) {
	if len(a.usages) == 0 {
		return nil, a.diagnostics
	}
	types := make(map[string]value.Type, len(a.usages))
	for name, uses := range a.usages {
		var typeCount [value.TypeFunction]uint
		for _, u := range uses {
			typeCount[u.Type]++
		}

		var maxType value.Type
		var maxCount uint
		for t, c := range typeCount {
			if c > maxCount {
				maxType, maxCount = value.Type(t), c
			}
		}

		types[name] = maxType

		for _, u := range uses {
			if u.Type != maxType {
				a.addDiagnostic(WrongUsage{
					WantType: u.Type,
					GotType:  maxType,
					Pos_:     u.Pos,
					Path_:    u.Path,
				})
			}
		}
	}
	return types, a.diagnostics
}
