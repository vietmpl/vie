package parser_test

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/vietmpl/vie/ast"
	"github.com/vietmpl/vie/parser"
)

func TestParseBytes(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	type testCase struct {
		file *File
	}

	unclosed := testCase{
		file: &File{
			Stmts: []Stmt{
				&Text{
					Value: "TOP\n",
				},
				&BadStmt{
					From: Pos{
						Line:      1,
						Character: 0,
					},
					To: Pos{
						Line:      3,
						Character: 0,
					},
				},
			},
		},
	}

	cases := map[string]testCase{
		"empty": {
			file: &File{
				Stmts: nil,
			},
		},
		"empty-tag":           unclosed,
		"unclosed-comm":       unclosed,
		"unclosed-render":     unclosed,
		"unclosed-render-var": unclosed,
		"nested-renders":      unclosed,
		"unclosed-switch-tag": unclosed,
		"unclosed-if-tag":     unclosed,
		"unclosed-end-tag":    unclosed,
		"unclosed-string-lit": unclosed,
		"unclosed-if": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&BadStmt{
						From: Pos{
							Line:      1,
							Character: 0,
						},
						To: Pos{
							Line:      3,
							Character: 0,
						},
					},
				},
			},
		},
		"unclosed-switch": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&BadStmt{
						From: Pos{
							Line:      1,
							Character: 0,
						},
						To: Pos{
							Line:      2,
							Character: 0,
						},
					},
				},
			},
		},
		"unclosed-switch-case": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&BadStmt{
						From: Pos{
							Line:      1,
							Character: 0,
						},
						To: Pos{
							Line:      2,
							Character: 0,
						},
					},
				},
			},
		},
		// TODO(skewb1k): would be nice to emit BadExpr in this case.
		"empty-render": unclosed,
		"empty-switch": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&SwitchStmt{
						Value: &BadExpr{
							From: Pos{
								Line:      1,
								Character: 9,
							},
							To: Pos{
								Line:      1,
								Character: 9,
							},
						},
						Cases: nil,
					},
					&Text{
						Value: "BOT\n",
					},
				},
			},
		},
		"empty-case": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&SwitchStmt{
						Value: &BasicLit{
							ValuePos: Pos{
								Line:      1,
								Character: 10,
							},
							Kind:  KindString,
							Value: "\"\"",
						},
						Cases: []CaseClause{
							{
								List: []Expr{
									&BadExpr{
										From: Pos{
											Line:      2,
											Character: 7,
										},
										To: Pos{
											Line:      2,
											Character: 7,
										},
									},
								},
								Body: nil,
							},
						},
					},
					&Text{
						Value: "BOT\n",
					},
				},
			},
		},
		"extra-case": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&SwitchStmt{
						Value: &BasicLit{
							ValuePos: Pos{
								Line:      1,
								Character: 10,
							},
							Kind:  KindString,
							Value: "\"\"",
						},
						Cases: []CaseClause{
							{
								List: []Expr{
									&BasicLit{
										ValuePos: Pos{
											Line:      2,
											Character: 8,
										},
										Kind:  KindString,
										Value: "''",
									},
									&BadExpr{
										From: Pos{
											Line:      2,
											Character: 10,
										},
										To: Pos{
											Line:      2,
											Character: 14,
										},
									},
								},
								Body: nil,
							},
						},
					},
					&Text{
						Value: "BOT\n",
					},
				},
			},
		},
		"empty-if": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&IfStmt{
						Cond: &BadExpr{
							From: Pos{
								Line:      1,
								Character: 5,
							},
							To: Pos{
								Line:      1,
								Character: 5,
							},
						},
						Cons:    nil,
						ElseIfs: nil,
						Else:    nil,
					},
					&Text{
						Value: "BOT\n",
					},
				},
			},
		},
		"empty-else-if": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&BadStmt{
						From: Pos{
							Line:      1,
							Character: 0,
						},
						To: Pos{
							Line:      4,
							Character: 0,
						},
					},
					&Text{
						Value: "BOT\n",
					},
				},
			},
		},
		"extra-else-if": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&IfStmt{
						Cond: &BasicLit{
							ValuePos: Pos{
								Line:      1,
								Character: 6,
							},
							Kind:  KindBool,
							Value: "true",
						},
						Cons: nil,
						ElseIfs: []ElseIfClause{
							{
								Cond: &BadExpr{
									From: Pos{
										Line:      2,
										Character: 11,
									},
									To: Pos{
										Line:      2,
										Character: 17,
									},
								},
								Cons: nil,
							},
						},
						Else: nil,
					},
					&Text{
						Value: "BOT\n",
					},
				},
			},
		},
		"extra-else": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&BadStmt{
						From: Pos{
							Line:      1,
							Character: 0,
						},
						To: Pos{
							Line:      4,
							Character: 0,
						},
					},
					&Text{
						Value: "BOT\n",
					},
				},
			},
		},
		"extra-render": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&RenderStmt{
						X: &BadExpr{
							From: Pos{
								Line:      1,
								Character: 5,
							},
							To: Pos{
								Line:      1,
								Character: 9,
							},
						},
					},
					&Text{
						Value: "\nBOT\n",
					},
				},
			},
		},
		"extra-render-2": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&RenderStmt{
						X: &BadExpr{
							From: Pos{
								Line:      1,
								Character: 5,
							},
							To: Pos{
								Line:      1,
								Character: 12,
							},
						},
					},
					&Text{
						Value: "\nBOT\n",
					},
				},
			},
		},
		"trailing-end-tag": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&BadStmt{
						From: Pos{
							Line:      1,
							Character: 0,
						},
						To: Pos{
							Line:      2,
							Character: 0,
						},
					},
					&Text{
						Value: "BOT\n",
					},
				},
			},
		},
		"trailing-case-tag": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&BadStmt{
						From: Pos{
							Line:      1,
							Character: 0,
						},
						To: Pos{
							Line:      2,
							Character: 0,
						},
					},
					&Text{
						Value: "BOT\n",
					},
				},
			},
		},
		"render-after-switch-tag": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&BadStmt{
						From: Pos{
							Line:      1,
							Character: 0,
						},
						To: Pos{
							Line:      4,
							Character: 0,
						},
					},
					&Text{
						Value: "BOT\n",
					},
				},
			},
		},
		"multiline-render": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&BadStmt{
						From: Pos{
							Line:      1,
							Character: 0,
						},
						To: Pos{
							Line:      4,
							Character: 0,
						},
					},
				},
			},
		},
		"multiline-comm": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&BadStmt{
						From: Pos{
							Line:      1,
							Character: 0,
						},
						To: Pos{
							Line:      2,
							Character: 0,
						},
					},
					&Text{
						Value: "\nBOT\n",
					},
				},
			},
		},
		"invalid-identifier": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&RenderStmt{
						X: &BadExpr{
							From: Pos{
								Line:      1,
								Character: 2,
							},
							To: Pos{
								Line:      1,
								Character: 4,
							},
						},
					},
					&Text{
						Value: "\nBOT\n",
					},
				},
			},
		},
		"unclosed-call": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&RenderStmt{
						X: &BadExpr{
							From: Pos{
								Line:      1,
								Character: 8,
							},
							To: Pos{
								Line:      1,
								Character: 10,
							},
						},
					},
					&Text{
						Value: "\nBOT\n",
					},
				},
			},
		},
		"extra-call": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&RenderStmt{
						X: &CallExpr{
							Func: Ident{
								NamePos: Pos{
									Line:      1,
									Character: 3,
								},
								Name: "@func",
							},
							Args: []Expr{
								&BasicLit{
									ValuePos: Pos{
										Line:      1,
										Character: 9,
									},
									Kind:  KindString,
									Value: "\"\"",
								},
								&BadExpr{
									From: Pos{
										Line:      1,
										Character: 11,
									},
									To: Pos{
										Line:      1,
										Character: 14,
									},
								},
							},
						},
					},
					&Text{
						Value: "\nBOT\n",
					},
				},
			},
		},
		"empty-pipe": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&RenderStmt{
						X: &BadExpr{
							From: Pos{
								Line:      1,
								Character: 2,
							},
							To: Pos{
								Line:      1,
								Character: 5,
							},
						},
					},
					&Text{
						Value: "\nBOT\n",
					},
				},
			},
		},
		"empty-pipe-2": {
			file: &File{
				Stmts: []Stmt{
					&Text{
						Value: "TOP\n",
					},
					&RenderStmt{
						X: &BadExpr{
							From: Pos{
								Line:      1,
								Character: 6,
							},
							To: Pos{
								Line:      1,
								Character: 8,
							},
						},
					},
					&Text{
						Value: "\nBOT\n",
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
			f, err := parser.ParseBytes(input, "")
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, cases[name].file, f)
		})
	}
}

// TestParseBytesFuzzTruncate checks that ParseBytes never panics on incomplete
// or slightly modified input. It concatenates all files in "testdata", then
// tests truncating from the start, truncating from the end, and removing each
// character.
func TestParseBytesFuzzTruncate(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)

	var combined []byte
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		combined = append(combined, b...)
	}

	src := combined

	// Truncate from start
	for i := range src {
		t.Run("start_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fragment := src[i:]
			_, err := parser.ParseBytes(fragment, "")
			if err != nil {
				t.Fatalf("unexpected error at length %d: %v", i, err)
			}
		})
	}

	// Truncate from end
	for i := len(src); i >= 0; i-- {
		t.Run("end_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fragment := src[:i]
			_, err := parser.ParseBytes(fragment, "")
			if err != nil {
				t.Fatalf("unexpected error at length %d: %v", i, err)
			}
		})
	}

	// Remove one character at each position
	for i := range src {
		t.Run("remove_"+strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			fragment := append([]byte(nil), src[:i]...)
			fragment = append(fragment, src[i+1:]...)
			_, err := parser.ParseBytes(fragment, "")
			if err != nil {
				t.Fatalf("unexpected error after removing index %d: %v", i, err)
			}
		})
	}
}
