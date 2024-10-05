package property

import (
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"nocomment/internal/filter"
	"nocomment/internal/langs"
)

func TestGoTokenRoundTrip(t *testing.T) {
	l, ok := langs.Default().ByName("go")
	if !ok {
		t.Fatal("go not registered")
	}
	src := []byte("package main\n\n// c\nfunc main() { /* d */ x := \"// e\"\n}\n")
	out, err := filter.Strip(l, src)
	if err != nil {
		t.Fatalf("strip: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "x.go", out, 0); err != nil {
		t.Fatalf("stripped source does not parse: %v", err)
	}
	if !slices.Equal(tokenStream(t, src), tokenStream(t, out)) {
		t.Fatalf("token stream changed")
	}
}

func TestGoTokenRoundTripCorpus(t *testing.T) {
	l, ok := langs.Default().ByName("go")
	if !ok {
		t.Fatal("go not registered")
	}
	dir := filepath.Join("..", "conformance", "testdata", "files", "go")
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Skip("corpus not fetched; run make corpus")
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(dir, file.Name())
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		out, err := filter.Strip(l, src)
		if err != nil {
			t.Fatalf("strip %s: %v", path, err)
		}
		if _, err := parser.ParseFile(token.NewFileSet(), path, out, 0); err != nil {
			t.Fatalf("stripped %s does not parse: %v", path, err)
		}
		if !slices.Equal(tokenStream(t, src), tokenStream(t, out)) {
			t.Fatalf("%s: token stream changed", path)
		}
	}
}

func tokenStream(t *testing.T, src []byte) []string {
	fset := token.NewFileSet()
	f := fset.AddFile("x.go", fset.Base(), len(src))
	var s scanner.Scanner
	s.Init(f, src, nil, 0)
	var out []string
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		out = append(out, tok.String())
		out = append(out, lit)
	}
	return out
}
