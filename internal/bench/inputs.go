package bench

import "nocomment/internal/corpus"

type Case struct {
	Language string
	Source   string
}

func Inputs() []Case {
	seen := map[string]bool{}
	out := make([]Case, 0, len(corpus.Cases()))
	for _, tc := range corpus.Cases() {
		if seen[tc.Language] {
			continue
		}
		seen[tc.Language] = true
		out = append(out, Case{Language: tc.Language, Source: tc.Input})
	}
	return out
}
