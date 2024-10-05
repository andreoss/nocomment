package golang

import (
	"strings"

	"nocomment/internal/lexer"
)

type Language struct{}

func New() Language { return Language{} }

func (Language) Name() string { return "go" }

func (Language) Extensions() []string { return []string{".go"} }

func (Language) Tokenize(src []byte) ([]lexer.Token, error) { return lexer.Tokenize(src) }

func (Language) IsComment(name string) bool {
	return strings.Contains(name, "COMMENT")
}
