package template

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

func (t Template) Render(context map[string]value.Value) (map[string][]byte, error) {
	files := make(map[string][]byte)

	var buf bytes.Buffer

	var walkFiles func(f []*File, parentPath string) error
	walkFiles = func(f []*File, parentPath string) error {
		for _, file := range f {
			render.MustRenderFile(&buf, file.NameAST, context)
			name := buf.String()
			buf.Reset()
			path := filepath.Join(parentPath, name)

			if file.ContentAST != nil {
				path = strings.TrimSuffix(path, ".vie")
				render.MustRenderFile(&buf, file.ContentAST, context)
				if _, exists := files[path]; exists {
					return fmt.Errorf("%s conflicts", path)
				}
				files[path] = buf.Bytes()
				buf.Reset()
			} else {
				files[path] = file.Content
			}
		}
		return nil
	}

	var walkDirs func(dirs []*Dir, parentPath string) error
	walkDirs = func(dirs []*Dir, parentPath string) error {
		for _, dir := range dirs {
			render.MustRenderFile(&buf, dir.NameAST, context)
			name := buf.String()
			buf.Reset()
			path := filepath.Join(parentPath, name)

			if err := walkFiles(dir.Files, path); err != nil {
				return err
			}
			if err := walkDirs(dir.Dirs, path); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walkFiles(t.Files, ""); err != nil {
		return nil, err
	}
	if err := walkDirs(t.Dirs, ""); err != nil {
		return nil, err
	}
	return files, nil
}
