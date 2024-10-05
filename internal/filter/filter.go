package filter

import (
	"bytes"

	"nocomment/internal/lang"
	"nocomment/internal/lexer"
)

func Strip(l lang.Language, src []byte) ([]byte, error) {
	tokens, err := l.Tokenize(src)
	if err != nil {
		return nil, err
	}
	if end := ShebangEnd(src); end > 0 {
		kept := make([]lexer.Token, 0, len(tokens))
		for _, t := range tokens {
			if t.Start < end {
				continue
			}
			kept = append(kept, t)
		}
		tokens = kept
	}
	return RemoveComments(src, tokens, l.IsComment), nil
}

func ShebangEnd(src []byte) int {
	if len(src) < 2 || src[0] != '#' || src[1] != '!' {
		return 0
	}
	if i := bytes.IndexByte(src, '\n'); i >= 0 {
		return i
	}
	return len(src)
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
		for end > t.Start && (src[end-1] == '\n' || src[end-1] == '\r') {
			end--
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
