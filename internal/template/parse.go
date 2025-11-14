package template

import (
	"os"
	"path/filepath"

	"github.com/vietmpl/vie/parser"
)

func FromDir(path string) (*Template, error) {
	parent := filepath.Dir(path)
	name := filepath.Base(path)

	rootDir, err := parseDir(parent, name)
	if err != nil {
		return nil, err
	}
	return &Template{
		Name:  name,
		Files: rootDir.Files,
		Dirs:  rootDir.Dirs,
	}, nil
}

func parseDir(parent, dirName string) (*Dir, error) {
	dirPath := filepath.Join(parent, dirName)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	// TODO(skewb1k): consider preallocating .Dirs and .Files.
	dir := Dir{
		Name: dirName,
	}

	for _, entry := range entries {
		name := entry.Name()
		// TODO(skewb1k): allow only specific subset of syntax in name.
		nameAST, err := parser.ParseBytes([]byte(name))
		if err != nil {
			return nil, err
		}
		if entry.IsDir() {
			subDir, err := parseDir(dirPath, name)
			if err != nil {
				return nil, err
			}
			// Parse nameAST only in subdirs.
			subDir.NameAST = nameAST
			dir.Dirs = append(dir.Dirs, subDir)
		} else {
			path := filepath.Join(dirPath, name)
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			f := &File{
				Name:    name,
				NameAST: nameAST,
				Content: content,
			}
			f.Content = content
			// Parse file content if its Vie file
			if filepath.Ext(name) == ".vie" {
				contentAST, err := parser.ParseBytes(content)
				if err != nil {
					return nil, err
				}
				f.ContentAST = contentAST
			}
			dir.Files = append(dir.Files, f)
		}
	}
	return &dir, nil
}
