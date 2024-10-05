package conformance

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"nocomment/internal/filter"
	"nocomment/internal/langs"
)

func TestRealWorldCorpus(t *testing.T) {
	languages, err := os.ReadDir(filepath.Join("testdata", "files"))
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
			dir := filepath.Join("testdata", "files", language)
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
				tokens, err := l.Tokenize(out)
				if err != nil {
					t.Fatalf("tokenize %s: %v", path, err)
				}
				protect := filter.ShebangEnd(out)
				for _, tok := range tokens {
					if l.IsComment(tok.Name) && tok.Start >= protect {
						t.Fatalf("%s: comment remains: %q", path, tok.Text)
					}
				}
			}
		})
	}
}

func TestManifest(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "manifest.tsv"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	registry := langs.Default()
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		t.Fatal("manifest is empty")
	}
	for _, line := range lines[1:] {
		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			t.Fatalf("bad manifest row: %q", line)
		}
		if _, ok := registry.ByName(fields[0]); !ok {
			t.Fatalf("%s not registered", fields[0])
		}
		for i, field := range fields {
			if strings.TrimSpace(field) == "" {
				t.Fatalf("row %q field %d empty", line, i)
			}
		}
	}
}

func isSubsequence(sub, src string) bool {
	i := 0
	for j := 0; j < len(src) && i < len(sub); j++ {
		if sub[i] == src[j] {
			i++
		}
	}
	return i == len(sub)
}
