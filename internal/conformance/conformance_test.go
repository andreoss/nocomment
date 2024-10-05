package conformance

import (
	"testing"

	"nocomment/internal/corpus"
	"nocomment/internal/filter"
	"nocomment/internal/langs"
)

func TestCommentRemoval(t *testing.T) {
	registry := langs.Default()
	for _, tc := range corpus.Cases() {
		t.Run(tc.Language, func(t *testing.T) {
			l, ok := registry.ByName(tc.Language)
			if !ok {
				t.Fatalf("%s (%s) not registered", tc.Language, tc.ID)
			}
			got, err := filter.Strip(l, []byte(tc.Input))
			if err != nil {
				t.Fatalf("Strip: %v", err)
			}
			if string(got) != tc.Expected {
				t.Fatalf("wrong result\n got: %q\nwant: %q", got, tc.Expected)
			}
		})
	}
}

func TestCasesAreComplete(t *testing.T) {
	for _, tc := range corpus.Cases() {
		if tc.ID == "" || tc.Language == "" {
			t.Fatalf("incomplete case: %+v", tc)
		}
		if tc.Input == tc.Expected {
			t.Fatalf("%s: fixture removes nothing", tc.Language)
		}
	}
}

func TestAllCasesRegistered(t *testing.T) {
	registry := langs.Default()
	for _, tc := range corpus.Cases() {
		if _, ok := registry.ByName(tc.Language); !ok {
			t.Fatalf("%s (%s) not registered", tc.Language, tc.ID)
		}
	}
}
