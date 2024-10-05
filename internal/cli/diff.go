package cli

import "github.com/pmezard/go-difflib/difflib"

func UnifiedDiff(name string, before, after []byte) []byte {
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(before)),
		B:        difflib.SplitLines(string(after)),
		FromFile: name,
		ToFile:   name,
		Context:  3,
	})
	if err != nil {
		return nil
	}
	return []byte(diff)
}
