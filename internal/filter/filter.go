package filter

import (
	"nocomment/internal/lang"
	"nocomment/internal/lexer"
)

func Strip(l lang.Language, src []byte) ([]byte, error) {
	tokens, err := l.Tokenize(src)
	if err != nil {
		return nil, err
	}
	return RemoveComments(src, tokens, l.IsComment), nil
}

func RemoveComments(src []byte, tokens []lexer.Token, isComment func(string) bool) []byte {
	out := make([]byte, 0, len(src))
	last := 0
	for _, t := range tokens {
		if !isComment(t.Name) {
			continue
		}
		if t.Start > len(src) {
			continue
		}
		end := t.Stop + 1
		if end < t.Start || end > len(src) {
			end = t.Start
		}
		if t.Start < last {
			if end > last {
				last = end
			}
			continue
		}
		out = append(out, src[last:t.Start]...)
		if end > last {
			last = end
		}
	}
	if last < len(src) {
		out = append(out, src[last:]...)
	}
	return out
}
