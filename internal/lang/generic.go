package lang

import "nocomment/internal/lexer"

type Generic struct {
	name     string
	exts     []string
	tokenize func([]byte) ([]lexer.Token, error)
	comments map[string]bool
}

func NewGeneric(name string, exts []string, tokenize func([]byte) ([]lexer.Token, error), comments []string) Generic {
	set := make(map[string]bool, len(comments))
	for _, c := range comments {
		set[c] = true
	}
	return Generic{name: name, exts: exts, tokenize: tokenize, comments: set}
}

func (g Generic) Name() string { return g.name }

func (g Generic) Extensions() []string { return g.exts }

func (g Generic) Tokenize(src []byte) ([]lexer.Token, error) { return g.tokenize(src) }

func (g Generic) IsComment(name string) bool { return g.comments[name] }
