package property

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"

	"nocomment/internal/filter"
	"nocomment/internal/langs"
)

func TestCorpusProperties(t *testing.T) {
	root := filepath.Join("..", "conformance", "testdata", "files")
	languages, err := os.ReadDir(root)
	if err != nil {
		t.Skip("corpus not fetched; run make corpus")
	}
	registry := langs.Default()
	for _, entry := range languages {
		if !entry.IsDir() {
			continue
		}
		language := entry.Name()
		l, ok := registry.ByName(language)
		if !ok {
			t.Fatalf("%s not registered", language)
		}
		t.Run(language, func(t *testing.T) {
			dir := filepath.Join(root, language)
			files, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				path := filepath.Join(dir, file.Name())
				src, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read %s: %v", path, err)
				}
				if !utf8.Valid(src) {
					t.Skipf("%s is not UTF-8", path)
				}
				out, err := filter.Strip(l, src)
				if err != nil {
					t.Fatalf("strip %s: %v", path, err)
				}
				if !isSubsequence(string(out), string(src)) {
					t.Fatalf("%s: output is not a subsequence", path)
				}
				again, err := filter.Strip(l, out)
				if err != nil {
					t.Fatalf("second strip %s: %v", path, err)
				}
				if !bytes.Equal(out, again) {
					t.Fatalf("%s: not idempotent", path)
				}
			}
		})
	}
}
