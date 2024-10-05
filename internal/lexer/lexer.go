package lexer

import (
	"unicode/utf8"

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
	return Collect(src, lex.NextToken, lex.SymbolicNames), nil
}

func Collect(src []byte, next func() antlr4.Token, names []string) []Token {
	offsets := runeOffsets(src)
	tokens := make([]Token, 0, 64)
	for {
		tok := next()
		if tok.GetTokenType() == antlr4.TokenEOF {
			break
		}
		start := runeByte(offsets, tok.GetStart())
		stop := runeByte(offsets, tok.GetStop()+1) - 1
		if stop < start {
			stop = start - 1
		}
		tokens = append(tokens, Token{
			Type:    tok.GetTokenType(),
			Name:    symbolicName(names, tok.GetTokenType()),
			Channel: tok.GetChannel(),
			Line:    tok.GetLine(),
			Column:  tok.GetColumn(),
			Start:   start,
			Stop:    stop,
		})
	}
	return tokens
}

func runeOffsets(src []byte) []int {
	offsets := make([]int, 0, len(src)+1)
	for i := 0; i < len(src); {
		offsets = append(offsets, i)
		_, size := utf8.DecodeRune(src[i:])
		i += size
	}
	offsets = append(offsets, len(src))
	return offsets
}

func runeByte(offsets []int, runeIndex int) int {
	if runeIndex < 0 {
		return 0
	}
	if runeIndex >= len(offsets) {
		return offsets[len(offsets)-1]
	}
	return offsets[runeIndex]
}

func symbolicName(names []string, tokenType int) string {
	if tokenType >= 0 && tokenType < len(names) {
		return names[tokenType]
	}
	return ""
}
