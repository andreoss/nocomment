package lexer

import (
	antlr4 "github.com/antlr4-go/antlr/v4"

	"nocomment/internal/lexer/generated/golang"
)

type Token struct {
	Type    int
	Name    string
	Text    string
	Channel int
	Line    int
	Column  int
	Start   int
	Stop    int
}

func Tokenize(src []byte) ([]Token, error) {
	lex := golang.NewGoLexer(antlr4.NewInputStream(string(src)))
	return Collect(lex.NextToken, lex.SymbolicNames), nil
}

func Collect(next func() antlr4.Token, names []string) []Token {
	tokens := make([]Token, 0, 64)
	for {
		tok := next()
		if tok.GetTokenType() == antlr4.TokenEOF {
			break
		}
		tokens = append(tokens, Token{
			Type:    tok.GetTokenType(),
			Name:    symbolicName(names, tok.GetTokenType()),
			Text:    tok.GetText(),
			Channel: tok.GetChannel(),
			Line:    tok.GetLine(),
			Column:  tok.GetColumn(),
			Start:   tok.GetStart(),
			Stop:    tok.GetStop(),
		})
	}
	return tokens
}

func symbolicName(names []string, tokenType int) string {
	if tokenType >= 0 && tokenType < len(names) {
		return names[tokenType]
	}
	return ""
}
