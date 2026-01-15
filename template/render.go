package template

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vietmpl/vie/render"
	"github.com/vietmpl/vie/value"
)

func (t Template) Render(context map[string]value.Value) (map[string][]byte, error) {
	files := make(map[string][]byte)

	onFile := func(f *File, parent string) error {
		name, err := render.Template(f.NameTemplate, context)
		if err != nil {
			return err
		}
		path := filepath.Join(parent, string(name))

		if f.ContentTemplate != nil {
			path = strings.TrimSuffix(path, ".vie")
			content, err := render.Template(f.ContentTemplate, context)
			if err != nil {
				return err
			}
			if _, ok := files[path]; ok {
				return fmt.Errorf("%s conflicts", path)
			}
			files[path] = content
		} else {
			files[path] = f.Content
		}
		return nil
	}

	onDir := func(d *Dir, parent string) error {
		name, err := render.Template(d.NameTemplate, context)
		if err != nil {
			return err
		}
		d.Name = string(name)
		return nil
	}

	if err := t.Walk(onDir, onFile); err != nil {
		return nil, err
	}
	return files, nil
}
