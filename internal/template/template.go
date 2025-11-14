package template

import "github.com/vietmpl/vie/ast"

type Template struct {
	Name  string
	Files []*File
	Dirs  []*Dir
}

type Dir struct {
	Name    string
	NameAST *ast.File
	Files   []*File
	Dirs    []*Dir
}

type File struct {
	Name       string
	NameAST    *ast.File
	Content    []byte
	ContentAST *ast.File
}
