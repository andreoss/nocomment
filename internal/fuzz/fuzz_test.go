package fuzz

import (
	"os"
	"path/filepath"
	"testing"

	"nocomment/internal/corpus"
	"nocomment/internal/filter"
	"nocomment/internal/langs"
)

func FuzzStrip(f *testing.F) {
	for _, tc := range corpus.Cases() {
		f.Add(tc.Language, tc.Input)
	}
	seedFromCorpus(f)
	f.Fuzz(func(t *testing.T, language, src string) {
		registry := langs.Default()
		l, ok := registry.ByName(language)
		if !ok {
			t.Skip()
		}
		if _, err := filter.Strip(l, []byte(src)); err != nil {
			t.Fatalf("Strip: %v", err)
		}
	})
}

func seedFromCorpus(f *testing.F) {
	root := filepath.Join("..", "conformance", "testdata", "files")
	languages, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, language := range languages {
		if !language.IsDir() {
			continue
		}
		dir := filepath.Join(root, language.Name())
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			src, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				continue
			}
			f.Add(language.Name(), string(src))
		}
	}
}
