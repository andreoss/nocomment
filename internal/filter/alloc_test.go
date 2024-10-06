package filter

import (
	"strings"
	"testing"

	"nocomment/internal/langs"
)

func allocsFor(t *testing.T, language string, lines int) float64 {
	t.Helper()
	l, ok := langs.Default().ByName(language)
	if !ok {
		t.Fatalf("%s not registered", language)
	}
	src := []byte(strings.Repeat("x = 1  # drop\n# whole line\n", lines/2))
	if _, err := Strip(l, src); err != nil {
		t.Fatalf("Strip: %v", err)
	}
	return testing.AllocsPerRun(3, func() { Strip(l, src) })
}

func TestStripDoesNotAllocatePerLine(t *testing.T) {
	small := allocsFor(t, "python", 1000)
	large := allocsFor(t, "python", 4000)
	if growth := large - small; growth > 32 {
		t.Fatalf("allocations grow with input: %.0f for 1000 lines, %.0f for 4000, growth %.0f; "+
			"a per-line allocation would grow by about 3000",
			small, large, growth)
	}
}
