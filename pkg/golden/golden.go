// Copyright 2025 skewb1k <skewb1kunix@gmail.com>. Licensed under the MIT License.

// Package golden provides utilities for running "golden" tests in Go.
//
// Golden tests are a technique for verifying that the output of your code
// matches expected results stored in "golden" files. This package automates
// reading input files, comparing their processed output to golden files, and
// updating the golden files.
//
// Stable tests verify that applying a transformation function does not change
// the input data. They are useful for checking that normalization or
// formatting operations are stable and idempotent.
//
// Directory structure example:
//
// testdata/
//
//	golden/
//	   test1.txt
//	   test1.txt.golden
//	   subdir/
//	     test2.txt
//	     test2.txt.golden
//	stable/
//	  test1.txt
//	  test2.txt
//
// All non-golden files under the "testdata" directory (including
// subdirectories) are treated as test inputs. For golden tests, each input
// file should have a corresponding golden file with the same name plus the
// ".golden" suffix. Files under the "stable" directory are used for stable
// tests, where the output of the tested function is expected to be identical
// to the input.
package golden

import (
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update golden files")

// Run executes golden tests for all files in the "testdata/golden"
// directory.
//
// It recursively scans the directory for all regular files that do not end
// with ".golden". For each input file, it finds or generates a corresponding
// golden file with the same name plus a ".golden" suffix.
//
// The function `f` is invoked for each test case with the testing.T and the
// input file contents. Each test runs as a subtest named after the relative
// file path and executes in parallel.
//
// If the -update flag is provided, the expected golden files are automatically
// overwritten with the newly generated output from `f`.
func Run(t *testing.T, f func(t *testing.T, input []byte) []byte) {
	run(t, "golden", false, f)
}

// RunStable verifies that applying `f` to the input data produces identical
// output.
//
// It scans all regular files in the specified subdirectory of the base
// "testdata" directory and checks that f(input) returns the same bytes as
// input.
//
// If the -update flag is set, the input files are overwritten with the output
// from `f`.
func RunStable(t *testing.T, f func(t *testing.T, input []byte) []byte) {
	run(t, "stable", true, f)
}

// RunGoldenTestdata is a backward-compatible wrapper for running golden tests
// on the base "testdata" directory.
func RunGoldenTestdata(t *testing.T, f func(t *testing.T, input []byte) []byte) {
	run(t, "", false, f)
}

const (
	baseDir      = "testdata"
	goldenSuffix = ".golden"
)

func run(t *testing.T, dir string, stable bool, f func(t *testing.T, input []byte) []byte) {
	t.Helper()

	testDir := filepath.Join(baseDir, dir)
	err := filepath.WalkDir(testDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip non-files and golden files
		if !d.Type().IsRegular() {
			return nil
		}

		if strings.HasSuffix(d.Name(), goldenSuffix) {
			if stable {
				t.Logf("warning: unexpected golden file %q found in \"stable\" test directory", path)
			}
			return nil
		}

		name, err := filepath.Rel(testDir, path)
		if err != nil {
			return err
		}

		t.Run(name, func(t *testing.T) {
			t.Helper()

			input, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			actual := f(t, input)

			goldenPath := path
			expected := input
			if !stable {
				goldenPath = goldenPath + goldenSuffix

				expected, err = os.ReadFile(goldenPath)
				if err != nil && !*update {
					t.Fatalf("missing golden file for %s: %v", path, err)
				}
			}

			if *update {
				if err := os.WriteFile(goldenPath, actual, 0644); err != nil {
					t.Fatalf("updating golden file: %v", err)
				}
				return
			}

			assert.Equal(t, string(expected), string(actual))
		})

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
