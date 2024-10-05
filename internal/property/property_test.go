package property

import (
	"bytes"
	"testing"

	"nocomment/internal/corpus"
	"nocomment/internal/filter"
	"nocomment/internal/langs"
)

func TestStripIsIdempotent(t *testing.T) {
	registry := langs.Default()
	for _, tc := range corpus.Cases() {
		t.Run(tc.Language, func(t *testing.T) {
			l, ok := registry.ByName(tc.Language)
			if !ok {
				t.Fatalf("%s not registered", tc.Language)
			}
			once, err := filter.Strip(l, []byte(tc.Input))
			if err != nil {
				t.Fatalf("first strip: %v", err)
			}
			twice, err := filter.Strip(l, once)
			if err != nil {
				t.Fatalf("second strip: %v", err)
			}
			if !bytes.Equal(once, twice) {
				t.Fatalf("not idempotent\n once: %q\n twice: %q", once, twice)
			}
		})
	}
}

func TestStripIsSubsequence(t *testing.T) {
	registry := langs.Default()
	for _, tc := range corpus.Cases() {
		t.Run(tc.Language, func(t *testing.T) {
			l, ok := registry.ByName(tc.Language)
			if !ok {
				t.Fatalf("%s not registered", tc.Language)
			}
			out, err := filter.Strip(l, []byte(tc.Input))
			if err != nil {
				t.Fatalf("strip: %v", err)
			}
			if !isSubsequence(string(out), tc.Input) {
				t.Fatalf("%q not a subsequence of %q", out, tc.Input)
			}
		})
	}
}

func TestStrippedHasNoComments(t *testing.T) {
	registry := langs.Default()
	for _, tc := range corpus.Cases() {
		t.Run(tc.Language, func(t *testing.T) {
			l, ok := registry.ByName(tc.Language)
			if !ok {
				t.Fatalf("%s not registered", tc.Language)
			}
			out, err := filter.Strip(l, []byte(tc.Input))
			if err != nil {
				t.Fatalf("strip: %v", err)
			}
			tokens, err := l.Tokenize(out)
			if err != nil {
				t.Fatalf("tokenize: %v", err)
			}
			protect := filter.ShebangEnd(out)
			for _, tok := range tokens {
				if l.IsComment(tok.Name) && tok.Start >= protect {
					t.Fatalf("comment remains after strip at %d: %s", tok.Start, tok.Name)
				}
			}
		})
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
