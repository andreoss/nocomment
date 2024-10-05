package fuzz

import (
	"testing"

	"nocomment/internal/corpus"
	"nocomment/internal/filter"
	"nocomment/internal/langs"
)

func FuzzStrip(f *testing.F) {
	for _, tc := range corpus.Cases() {
		f.Add(tc.Language, tc.Input)
	}
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
