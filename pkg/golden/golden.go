// Copyright 2025 skewb1k <skewb1kunix@gmail.com>. Licensed under the MIT License.

// Package golden provides utilities for running "golden" tests in Go.
//
// Golden tests are a technique for verifying that the output of your code
// matches expected results stored in "golden" files. This package automates
// reading input files, comparing their processed output to golden files, and
// updating the golden files.
package golden

import (
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "update golden files")

const (
	dir          = "testdata"
	goldenSuffix = ".golden"
	stableSuffix = ".stable"
)

// Run executes golden tests for all files in the "testdata" directory.
//
// It recursively scans the directory for all regular files that do not end
// with ".golden". For each input file, it looks for a corresponding golden
// file with the same name plus the ".golden" suffix. The provided function `f`
// is called for each test case with the testing.T and the input file content.
// Each test runs as a subtest named after the relative file path.
//
// Also, it supports "stable" tests. Stable tests are a special kind of golden
// test where the input and output are expected to match exactly. This is
// useful for verifying operations such as formatting or normalization that are
// meant to be idempotent. Stable test files should have the ".stable" suffix
// and do not require a separate golden file.
//
// When the `-update` flag is set, golden files are overwritten with the output
// produced by `f`. If the corresponding golden file does not exist, it will be
// created. For stable tests, the original input file will be overwritten.
//
// Example directory structure:
//
//	testdata/
//	  example1.txt
//	  example1.txt.golden
//	  example2.txt.stable
//	  subdir/
//	    example3.txt
//	    example3.txt.golden
//
// Example usage:
//
//	func TestFormat(t *testing.T) {
//	    golden.Run(t, func(t *testing.T, input []byte) []byte {
//	        formatted, err := format.Source(input)
//	        if err != nil {
//	            t.Fatalf("failed to format Go source: %v", err)
//	        }
//	        return formatted
//	    })
//	}
func Run(t *testing.T, f func(t *testing.T, input []byte) []byte) {
	t.Helper()
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip non-files and golden files
		if !d.Type().IsRegular() || strings.HasSuffix(d.Name(), goldenSuffix) {
			return nil
		}

		name, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		t.Run(name, func(t *testing.T) {
			t.Helper()

			input, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			want := input
			expectedPath := path

			stable := strings.HasSuffix(path, stableSuffix)
			if !stable {
				expectedPath = path + goldenSuffix
				want, err = os.ReadFile(expectedPath)
				if err != nil && !*update {
					t.Fatalf("missing golden file for %s: %v", path, err)
				}
			}

			got := f(t, input)

			if *update {
				if err := os.WriteFile(expectedPath, got, 0644); err != nil {
					t.Fatalf("updating golden file: %v", err)
				}
				return
			}

			if diff := cmp.Diff(string(want), string(got)); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
