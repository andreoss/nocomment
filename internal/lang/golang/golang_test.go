package golang

import "testing"

func TestLanguage(t *testing.T) {
	l := New()
	if l.Name() != "go" {
		t.Fatalf("name = %q", l.Name())
	}
	if exts := l.Extensions(); len(exts) != 1 || exts[0] != ".go" {
		t.Fatalf("extensions = %v", exts)
	}
	for _, name := range []string{"COMMENT", "LINE_COMMENT", "COMMENT_NLSEMI", "LINE_COMMENT_NLSEMI"} {
		if !l.IsComment(name) {
			t.Fatalf("IsComment(%q) = false", name)
		}
	}
	if l.IsComment("IDENTIFIER") {
		t.Fatal("IsComment(IDENTIFIER) = true")
	}
	tokens, err := l.Tokenize([]byte("package main // c\n"))
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(tokens) == 0 {
		t.Fatal("no tokens")
	}
}
