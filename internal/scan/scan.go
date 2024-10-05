package scan

import "bytes"

type StringRule struct {
	Open      string
	Close     string
	Escape    bool
	Doubled   bool
	Multiline bool
}

type Syntax struct {
	Line           []string
	Block          [][2]string
	BlockLineStart [][2]string
	Nest           bool
	LineBoundary   bool
	Strings        []StringRule
	Regex          bool
	Percent        bool
	HereDocs       string
	YAMLBlock      bool
}

type Range struct {
	Start int
	Stop  int
}

var regexKeywords = map[string]bool{
	"return": true, "typeof": true, "instanceof": true, "in": true,
	"of": true, "new": true, "delete": true, "void": true, "case": true,
	"do": true, "else": true, "yield": true, "await": true,
}

func Ranges(src []byte, s Syntax) []Range {
	out := make([]Range, 0, 8)
	i := 0
	for i < len(src) {
		if start, end, ok := matchLine(src, i, s); ok {
			out = append(out, Range{Start: start, Stop: end})
			i = end
			continue
		}
		if open, close, ok := matchBlock(src, i, s.Block, false); ok {
			start := i
			i += len(open)
			depth := 1
			for i < len(src) && depth > 0 {
				if s.Nest && hasPrefix(src, i, open) {
					depth++
					i += len(open)
					continue
				}
				if hasPrefix(src, i, close) {
					depth--
					i += len(close)
					continue
				}
				i++
			}
			out = append(out, Range{Start: start, Stop: i})
			continue
		}
		if s.Regex && src[i] == '/' && regexAllowed(src, i) {
			i = scanRegex(src, i)
			continue
		}
		if open, close, ok := matchBlock(src, i, s.BlockLineStart, true); ok {
			start := i
			i += len(open)
			for i < len(src) {
				if hasPrefix(src, i, close) && atLineStart(src, i) {
					i += len(close)
					break
				}
				i++
			}
			out = append(out, Range{Start: start, Stop: i})
			continue
		}
		if s.Percent && src[i] == '%' {
			if end, ok := scanPercent(src, i); ok {
				i = end
				continue
			}
		}
		if s.HereDocs == "shell" && hasPrefix(src, i, "<<") {
			if end, ok := scanHereDoc(src, i); ok {
				i = end
				continue
			}
		}
		if s.HereDocs == "php" && hasPrefix(src, i, "<<<") {
			if end, ok := scanPHPHereDoc(src, i); ok {
				i = end
				continue
			}
		}
		if s.YAMLBlock && (src[i] == '|' || src[i] == '>') {
			if end, ok := scanYAMLBlock(src, i); ok {
				i = end
				continue
			}
		}
		if rule, ok := matchString(src, i, s.Strings); ok {
			i = scanString(src, i, rule)
			continue
		}
		i++
	}
	return out
}

func matchLine(src []byte, i int, s Syntax) (int, int, bool) {
	for _, marker := range s.Line {
		if !hasPrefix(src, i, marker) {
			continue
		}
		if s.LineBoundary && !atBoundary(src, i) {
			continue
		}
		j := i + len(marker)
		for j < len(src) && src[j] != '\n' && src[j] != '\r' {
			j++
		}
		return i, j, true
	}
	return 0, 0, false
}

func matchBlock(src []byte, i int, pairs [][2]string, lineStart bool) (string, string, bool) {
	for _, pair := range pairs {
		if !hasPrefix(src, i, pair[0]) {
			continue
		}
		if lineStart && !atLineStart(src, i) {
			continue
		}
		return pair[0], pair[1], true
	}
	return "", "", false
}

func matchString(src []byte, i int, rules []StringRule) (StringRule, bool) {
	var best StringRule
	found := false
	for _, rule := range rules {
		if !hasPrefix(src, i, rule.Open) {
			continue
		}
		if !found || len(rule.Open) > len(best.Open) {
			best = rule
			found = true
		}
	}
	return best, found
}

func scanString(src []byte, i int, rule StringRule) int {
	i += len(rule.Open)
	for i < len(src) {
		if rule.Escape && src[i] == '\\' {
			i += 2
			continue
		}
		if rule.Doubled && hasPrefix(src, i, rule.Close) && hasPrefix(src, i+len(rule.Close), rule.Close) {
			i += 2 * len(rule.Close)
			continue
		}
		if hasPrefix(src, i, rule.Close) {
			return i + len(rule.Close)
		}
		if !rule.Multiline && (src[i] == '\n' || src[i] == '\r') {
			return i
		}
		i++
	}
	return i
}

func regexAllowed(src []byte, i int) bool {
	j := i - 1
	for j >= 0 && isSpace(src[j]) {
		j--
	}
	if j < 0 {
		return true
	}
	switch src[j] {
	case '(', ',', '=', ':', '[', '!', '&', '|', '?', '{', '}', ';', '+', '-', '*', '%', '<', '>', '^', '~':
		return true
	}
	if isWordByte(src[j]) {
		start := j
		for start >= 0 && isWordByte(src[start]) {
			start--
		}
		return regexKeywords[string(src[start+1:j+1])]
	}
	return false
}

func scanRegex(src []byte, i int) int {
	i++
	inClass := false
	for i < len(src) {
		c := src[i]
		if c == '\\' {
			i += 2
			continue
		}
		if c == '\n' || c == '\r' {
			return i
		}
		if c == '[' {
			inClass = true
		}
		if c == ']' {
			inClass = false
		}
		if c == '/' && !inClass {
			return i + 1
		}
		i++
	}
	return i
}

func scanPercent(src []byte, i int) (int, bool) {
	j := i + 1
	if j < len(src) && isPercentType(src[j]) {
		j++
	}
	if j >= len(src) {
		return i, false
	}
	open := src[j]
	if isAlnum(open) || isSpace(open) {
		return i, false
	}
	close := matchingDelim(open)
	nest := close != open
	depth := 1
	j++
	for j < len(src) {
		if src[j] == '\\' {
			j += 2
			continue
		}
		if nest && src[j] == open {
			depth++
			j++
			continue
		}
		if src[j] == close {
			depth--
			if depth == 0 {
				return j + 1, true
			}
		}
		j++
	}
	return j, true
}

func scanHereDoc(src []byte, i int) (int, bool) {
	if i > 0 && !isSpace(src[i-1]) && src[i-1] != ';' && src[i-1] != '&' && src[i-1] != '|' {
		return i, false
	}
	j := i + 2
	dash := false
	if j < len(src) && src[j] == '-' {
		dash = true
		j++
	}
	if j < len(src) && src[j] == '<' {
		return i, false
	}
	for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
		j++
	}
	quote := byte(0)
	if j < len(src) && (src[j] == '\'' || src[j] == '"') {
		quote = src[j]
		j++
	}
	start := j
	for j < len(src) && isWordByte(src[j]) {
		j++
	}
	if j == start {
		return i, false
	}
	if !isAlphaByte(src[start]) && src[start] != '_' {
		return i, false
	}
	delim := string(src[start:j])
	if quote != 0 && j < len(src) && src[j] == quote {
		j++
	}
	for j < len(src) && src[j] != '\n' {
		j++
	}
	if j < len(src) {
		j++
	}
	for j <= len(src) {
		lineEnd := j
		for lineEnd < len(src) && src[lineEnd] != '\n' {
			lineEnd++
		}
		line := src[j:lineEnd]
		if dash {
			line = bytes.TrimLeft(line, "\t")
		}
		if string(line) == delim {
			if lineEnd < len(src) {
				return lineEnd + 1, true
			}
			return lineEnd, true
		}
		if lineEnd >= len(src) {
			break
		}
		j = lineEnd + 1
	}
	return len(src), true
}

func scanPHPHereDoc(src []byte, i int) (int, bool) {
	j := i + 3
	for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
		j++
	}
	quote := byte(0)
	if j < len(src) && (src[j] == '\'' || src[j] == '"') {
		quote = src[j]
		j++
	}
	start := j
	for j < len(src) && isWordByte(src[j]) {
		j++
	}
	if j == start {
		return i, false
	}
	delim := string(src[start:j])
	if quote != 0 && j < len(src) && src[j] == quote {
		j++
	}
	for j < len(src) && src[j] != '\n' {
		j++
	}
	if j < len(src) {
		j++
	}
	for j <= len(src) {
		lineEnd := j
		for lineEnd < len(src) && src[lineEnd] != '\n' {
			lineEnd++
		}
		line := bytes.TrimLeft(src[j:lineEnd], " \t")
		if bytes.HasPrefix(line, []byte(delim)) {
			rest := line[len(delim):]
			if len(rest) == 0 || (len(rest) == 1 && rest[0] == ';') {
				if lineEnd < len(src) {
					return lineEnd + 1, true
				}
				return lineEnd, true
			}
		}
		if lineEnd >= len(src) {
			break
		}
		j = lineEnd + 1
	}
	return len(src), true
}

func scanYAMLBlock(src []byte, i int) (int, bool) {
	j := i + 1
	for j < len(src) && (src[j] == '+' || src[j] == '-' || (src[j] >= '0' && src[j] <= '9')) {
		j++
	}
	for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
		j++
	}
	if j < len(src) && src[j] != '\n' && src[j] != '\r' {
		return i, false
	}
	lineStart := i
	for lineStart > 0 && src[lineStart-1] != '\n' {
		lineStart--
	}
	indent := 0
	for k := lineStart; k < i; k++ {
		if src[k] == ' ' {
			indent++
		} else if src[k] == '\t' {
			indent += 8
		} else {
			break
		}
	}
	if j < len(src) && src[j] == '\r' {
		j++
	}
	if j < len(src) && src[j] == '\n' {
		j++
	}
	for j < len(src) {
		lineEnd := j
		for lineEnd < len(src) && src[lineEnd] != '\n' {
			lineEnd++
		}
		line := src[j:lineEnd]
		if len(bytes.TrimSpace(line)) == 0 {
			if lineEnd < len(src) {
				j = lineEnd + 1
				continue
			}
			return lineEnd, true
		}
		ind := 0
		for k := j; k < lineEnd; k++ {
			if src[k] == ' ' {
				ind++
			} else if src[k] == '\t' {
				ind += 8
			} else {
				break
			}
		}
		if ind <= indent {
			return j, true
		}
		if lineEnd >= len(src) {
			return lineEnd, true
		}
		j = lineEnd + 1
	}
	return j, true
}

func hasPrefix(src []byte, i int, prefix string) bool {
	if i < 0 || i+len(prefix) > len(src) {
		return false
	}
	for k := 0; k < len(prefix); k++ {
		if src[i+k] != prefix[k] {
			return false
		}
	}
	return true
}

func atLineStart(src []byte, i int) bool {
	return i == 0 || src[i-1] == '\n'
}

func atBoundary(src []byte, i int) bool {
	if i == 0 {
		return true
	}
	switch src[i-1] {
	case ' ', '\t', '\n', '\r', ';', '(', ')':
		return true
	}
	return false
}

func isPercentType(c byte) bool {
	switch c {
	case 'q', 'Q', 'w', 'W', 'i', 'I', 'r', 's', 'x':
		return true
	}
	return false
}

func matchingDelim(open byte) byte {
	switch open {
	case '(':
		return ')'
	case '[':
		return ']'
	case '{':
		return '}'
	case '<':
		return '>'
	}
	return open
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isAlphaByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isAlnum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func isWordByte(c byte) bool {
	return isAlnum(c) || c == '_' || c == '$'
}
