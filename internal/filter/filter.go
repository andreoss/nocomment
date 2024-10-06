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
	mask := commentMask(src, tokens, isComment)
	out := make([]byte, 0, len(src))
	for start := 0; start < len(src); {
		body, next := splitLine(src, start)
		if anyMasked(mask, start, body) && keptIsBlank(src, mask, start, body) &&
			!anyMasked(mask, body, next) {
			start = next
			continue
		}
		out = appendKept(out, src, mask, start, next)
		start = next
	}
	return out
}

func commentMask(src []byte, tokens []lexer.Token, isComment func(string) bool) []bool {
	mask := make([]bool, len(src))
	for _, t := range tokens {
		if !isComment(t.Name) || t.Start > len(src) {
			continue
		}
		end := t.Stop + 1
		if end < t.Start || end > len(src) {
			end = t.Start
		}
		for end > t.Start && (src[end-1] == '\n' || src[end-1] == '\r') {
			end--
		}
		for i := t.Start; i < end; i++ {
			if i >= 0 && i < len(mask) {
				mask[i] = true
			}
		}
	}
	return mask
}

func splitLine(src []byte, start int) (body, next int) {
	i := start
	for i < len(src) && src[i] != '\n' {
		i++
	}
	body = i
	if body > start && src[body-1] == '\r' {
		body--
	}
	if i < len(src) {
		return body, i + 1
	}
	return body, i
}

func appendKept(dst, src []byte, mask []bool, from, to int) []byte {
	for i := from; i < to; i++ {
		if !mask[i] {
			dst = append(dst, src[i])
		}
	}
	return dst
}

func anyMasked(mask []bool, from, to int) bool {
	for i := from; i < to; i++ {
		if mask[i] {
			return true
		}
	}
	return false
}

func keptIsBlank(src []byte, mask []bool, from, to int) bool {
	for i := from; i < to; i++ {
		if mask[i] {
			continue
		}
		b := src[i]
		if b != ' ' && b != '\t' && b != '\r' && b != '\v' && b != '\f' {
			return false
		}
	}
	return true
}
