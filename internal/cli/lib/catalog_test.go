package cli

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func createMapFS(paths []string) fstest.MapFS {
	fsys := make(fstest.MapFS)

	normalizePath := func(p string) string {
		p = strings.ReplaceAll(p, "\\", "/")
		p = strings.TrimPrefix(p, "./")
		return p
	}

	ensureDir := func(dirPath string) {
		dirPath = normalizePath(dirPath)
		if dirPath == "." || dirPath == "" {
			return
		}
		if _, exists := fsys[dirPath]; !exists {
			fsys[dirPath] = &fstest.MapFile{Mode: fs.ModeDir}
		}
	}

	for _, path := range paths {
		path = normalizePath(path)

		parts := strings.Split(path, "/")
		for i := 1; i <= len(parts); i++ {
			parentPath := strings.Join(parts[:i], "/")
			ensureDir(parentPath)
		}
	}

	return fsys
}

func TestSearchTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		structure []string
		baseDir   string
		query     string
		want      string
		wantEmpty bool
	}{
		{
			name:      "simple template in .vie",
			structure: []string{".vie/svo"},
			baseDir:   ".",
			query:     "svo",
			want:      ".vie/svo",
		},
		{
			name:      "named group template",
			structure: []string{".vie/[group]/svo"},
			baseDir:   ".",
			query:     "group/svo",
			want:      ".vie/[group]/svo",
		},
		{
			name:      "parentheses group template",
			structure: []string{".vie/(name)/gg"},
			baseDir:   ".",
			query:     "(name)/gg",
			want:      ".vie/(name)/gg",
		},
		{
			name:      "template in parentheses group (search by name)",
			structure: []string{".vie/(name)/gg"},
			baseDir:   ".",
			query:     "gg",
			want:      ".vie/(name)/gg",
		},
		{
			name:      "named group inside parentheses group",
			structure: []string{".vie/(name)/[group]/svo"},
			baseDir:   ".",
			query:     "group/svo",
			want:      ".vie/(name)/[group]/svo",
		},
		{
			name:      "nested named groups",
			structure: []string{".vie/[group]/[group]/svo"},
			baseDir:   ".",
			query:     "group/group/svo",
			want:      ".vie/[group]/[group]/svo",
		},
		{
			name:      "mixed groups: (group)/[group]/tmpl",
			structure: []string{".vie/(group)/[group]/tmpl"},
			baseDir:   ".",
			query:     "group/tmpl",
			want:      ".vie/(group)/[group]/tmpl",
		},
		{
			name:      "template in non-named group",
			structure: []string{".vie/mygroup/tmpl"},
			baseDir:   ".",
			query:     "tmpl",
			want:      ".vie/mygroup/tmpl",
		},
		{
			name:      "template not found",
			structure: []string{".vie"},
			baseDir:   ".",
			query:     "nonexistent",
			wantEmpty: true,
		},
		{
			name:      "nested parentheses groups",
			structure: []string{".vie/(group1)/(group2)/tmpl"},
			baseDir:   ".",
			query:     "tmpl",
			want:      ".vie/(group1)/(group2)/tmpl",
		},
		{
			name:      "multiple templates - find first",
			structure: []string{".vie/tmpl1", ".vie/group/tmpl1"},
			baseDir:   ".",
			query:     "tmpl1",
			want:      ".vie/tmpl1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if strings.HasPrefix(tt.query, "/") || strings.HasPrefix(tt.query, "./") || strings.HasPrefix(tt.query, "../") {
				result := SearchTemplate(nil, tt.baseDir, tt.query)
				assert.Equal(t, tt.query, result)
				return
			}

			fsys := createMapFS(tt.structure)

			result := SearchTemplate(fsys, tt.baseDir, tt.query)

			if tt.wantEmpty {
				assert.Empty(t, result, "Expected empty result")
			} else {
				want := filepath.Clean(tt.want)
				got := filepath.Clean(result)
				assert.Equal(t, want, got, "SearchTemplate returned unexpected path")
			}
		})
	}
}

func TestIsPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
		skipOS   string
	}{
		{"empty string", "", false, ""},
		{"current dir", ".", true, ""},
		{"relative path with ./", "./path", true, ""},
		{"relative path with ../", "../path", true, ""},
		{"absolute path unix", "/absolute/path", true, "windows"},
		{"absolute path windows", "C:\\path", true, "linux"},
		{"template name", "template", false, ""},
		{"group/template", "group/template", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.skipOS != "" && runtime.GOOS == tt.skipOS {
				t.Skipf("Skipping test on %s", runtime.GOOS)
			}

			result := isPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNamedGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"named group with brackets", "[group]", true},
		{"not named group with parentheses", "(group)", false},
		{"regular name", "group", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isNamedGroup(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
