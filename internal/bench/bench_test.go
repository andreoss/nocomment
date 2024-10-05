package bench

import (
	"strings"
	"testing"

	"nocomment/internal/filter"
	"nocomment/internal/langs"
)

func BenchmarkStrip(b *testing.B) {
	registry := langs.Default()
	for _, tc := range Inputs() {
		b.Run(tc.Language, func(b *testing.B) {
			l, ok := registry.ByName(tc.Language)
			if !ok {
				b.Fatalf("%s not registered", tc.Language)
			}
			src := []byte(strings.Repeat(tc.Source, 1000))
			b.SetBytes(int64(len(src)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := filter.Strip(l, src); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
