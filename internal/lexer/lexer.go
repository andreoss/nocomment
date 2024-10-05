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
}

func Tokenize(src []byte) ([]Token, error) {
	input := antlr4.NewInputStream(string(src))
	lex := golang.NewGoLexer(input)
	tokens := make([]Token, 0, len(src)/4+1)
	for {
		tok := lex.NextToken()
		if tok.GetTokenType() == antlr4.TokenEOF {
			break
		}
		tokens = append(tokens, Token{
			Type:    tok.GetTokenType(),
			Name:    symbolicName(lex.SymbolicNames, tok.GetTokenType()),
			Text:    tok.GetText(),
			Channel: tok.GetChannel(),
			Line:    tok.GetLine(),
			Column:  tok.GetColumn(),
		})
	}
	return tokens, nil
}

func symbolicName(names []string, tokenType int) string {
	if tokenType >= 0 && tokenType < len(names) {
		return names[tokenType]
	}
	return ""
}
