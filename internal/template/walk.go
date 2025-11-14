package template

import "path/filepath"

func (t Template) Walk(onDir func(*Dir, string) error, onFile func(*File, string) error) error {
	if err := walkFiles(t.Files, "", onFile); err != nil {
		return err
	}
	if err := walkDirs(t.Dirs, "", onDir, onFile); err != nil {
		return err
	}
	return nil
}

func walkDirs(
	dirs []*Dir,
	parent string,
	onDir func(*Dir, string) error,
	onFile func(*File, string) error,
) error {
	for _, d := range dirs {
		if err := onDir(d, parent); err != nil {
			return err
		}
		if err := walkFiles(d.Files, filepath.Join(parent, d.Name), onFile); err != nil {
			return err
		}
		if err := walkDirs(d.Dirs, filepath.Join(parent, d.Name), onDir, onFile); err != nil {
			return err
		}
	}
	return nil
}

func walkFiles(files []*File, parent string, onFile func(*File, string) error) error {
	for _, f := range files {
		if err := onFile(f, parent); err != nil {
			return err
		}
	}
	return nil
}
