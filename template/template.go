package template

import "github.com/vietmpl/vie/ast"

type Template struct {
	Name  string
	Files []*File
	Dirs  []*Dir
}

type Dir struct {
	Name         string
	NameTemplate *ast.Template
	Files        []*File
	Dirs         []*Dir
}

type File struct {
	Name            string
	NameTemplate    *ast.Template
	Content         []byte
	ContentTemplate *ast.Template
}
