package lang

import (
	"testing"

	"nocomment/internal/lexer"
)

func TestGeneric(t *testing.T) {
	g := NewGeneric("x", []string{".x"}, func([]byte) ([]lexer.Token, error) {
		return []lexer.Token{{Name: "COMMENT"}}, nil
	}, []string{"COMMENT", "LINE_COMMENT"})
	if g.Name() != "x" {
		t.Fatalf("name = %q", g.Name())
	}
	if len(g.Extensions()) != 1 || g.Extensions()[0] != ".x" {
		t.Fatalf("extensions = %v", g.Extensions())
	}
	if !g.IsComment("COMMENT") || !g.IsComment("LINE_COMMENT") {
		t.Fatal("comment names not matched")
	}
	if g.IsComment("IDENTIFIER") {
		t.Fatal("unexpected comment match")
	}
	tokens, err := g.Tokenize(nil)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	if len(tokens) != 1 || tokens[0].Name != "COMMENT" {
		t.Fatalf("tokens = %+v", tokens)
	}
}
