package bench

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"nocomment/internal/filter"
	"nocomment/internal/langs"
)

func TestAllocationBudget(t *testing.T) {
	baseline := readBaseline(t)
	registry := langs.Default()
	for _, tc := range Inputs() {
		l, ok := registry.ByName(tc.Language)
		if !ok {
			t.Fatalf("%s not registered", tc.Language)
		}
		want, ok := baseline[tc.Language]
		if !ok {
			t.Fatalf("no baseline for %s", tc.Language)
		}
		src := []byte(strings.Repeat(tc.Source, 1000))
		filter.Strip(l, src)
		allocs := testing.AllocsPerRun(3, func() {
			filter.Strip(l, src)
		})
		if allocs > want*1.10 {
			t.Fatalf("%s: allocations per run %.0f exceed baseline %.0f by more than 10 percent", tc.Language, allocs, want)
		}
	}
}

func readBaseline(t *testing.T) map[string]float64 {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "baseline.tsv"))
	if err != nil {
		t.Fatalf("open baseline: %v", err)
	}
	defer f.Close()
	out := map[string]float64{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "language") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			t.Fatalf("bad baseline row %q", line)
		}
		value, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			t.Fatalf("bad baseline value %q", fields[1])
		}
		out[fields[0]] = value
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	return out
}
