package scan

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
}

type Range struct {
	Start int
	Stop  int
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
