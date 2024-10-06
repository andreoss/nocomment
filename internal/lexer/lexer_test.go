package lexer

import (
	"strings"
	"testing"

	antlr4 "github.com/antlr4-go/antlr/v4"

	"nocomment/internal/lexer/generated/golang"
)

func tokenize(src []byte) ([]Token, error) {
	lex := golang.NewGoLexer(antlr4.NewInputStream(string(src)))
	return Collect(src, lex.NextToken, lex.SymbolicNames), nil
}

var commentNames = map[string]bool{
	"LINE_COMMENT":        true,
	"COMMENT":             true,
	"LINE_COMMENT_NLSEMI": true,
	"COMMENT_NLSEMI":      true,
}

func testText(src []byte, tok Token) string {
	if tok.Start < 0 || tok.Stop+1 > len(src) || tok.Stop < tok.Start-1 {
		return ""
	}
	return string(src[tok.Start : tok.Stop+1])
}

func TestTokenizeSeparatesCommentsAndStrings(t *testing.T) {
	src := []byte("package main\n// first\nvar s = \"// not a comment\"\n/* block */\nvar t = `/* raw */`\n")
	tokens, err := tokenize(src)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	comments := 0
	interpreted := 0
	raw := 0
	for _, tok := range tokens {
		if commentNames[tok.Name] {
			comments++
			if tok.Channel == 0 {
				t.Fatalf("comment %q on default channel", testText(src, tok))
			}
		}
		if tok.Name == "INTERPRETED_STRING_LIT" {
			interpreted++
			if testText(src, tok) != "\"// not a comment\"" {
				t.Fatalf("interpreted string = %q", testText(src, tok))
			}
		}
		if tok.Name == "RAW_STRING_LIT" {
			raw++
			if testText(src, tok) != "`/* raw */`" {
				t.Fatalf("raw string = %q", testText(src, tok))
			}
		}
	}
	if comments != 2 {
		t.Fatalf("comments = %d, tokens = %+v", comments, tokens)
	}
	if interpreted != 1 {
		t.Fatalf("interpreted strings = %d", interpreted)
	}
	if raw != 1 {
		t.Fatalf("raw strings = %d", raw)
	}
}

func TestTokenizeByteOffsetsWithMultibyte(t *testing.T) {
	src := []byte("package main\n// \u201cx\u201d\nvar s = 1\n")
	tokens, err := tokenize(src)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	found := false
	for _, tok := range tokens {
		if commentNames[tok.Name] {
			found = true
			if got := testText(src, tok); got != "// \u201cx\u201d" {
				t.Fatalf("comment bytes = %q", got)
			}
		}
	}
	if !found {
		t.Fatalf("no comment in %+v", tokens)
	}
}

func TestTokenizePositions(t *testing.T) {
	src := []byte("package main // c\n")
	tokens, err := tokenize(src)
	if err != nil {
		t.Fatalf("Tokenize: %v", err)
	}
	found := false
	for _, tok := range tokens {
		if commentNames[tok.Name] {
			found = true
			if strings.TrimSpace(testText(src, tok)) != "// c" {
				t.Fatalf("text = %q", testText(src, tok))
			}
			if tok.Line != 1 {
				t.Fatalf("line = %d", tok.Line)
			}
		}
	}
	if !found {
		t.Fatalf("no comment token in %+v", tokens)
	}
}

func TestSymbolicName(t *testing.T) {
	if got := symbolicName([]string{"", "A"}, 1); got != "A" {
		t.Fatalf("got %q", got)
	}
	if got := symbolicName([]string{"", "A"}, 5); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := symbolicName([]string{"", "A"}, -1); got != "" {
		t.Fatalf("got %q", got)
	}
}
