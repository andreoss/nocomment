package filter

import (
	"testing"

	"nocomment/internal/lang/golang"
	"nocomment/internal/langs"
	"nocomment/internal/lexer"
)

func isComment(name string) bool {
	return name == "COMMENT" || name == "LINE_COMMENT"
}

func TestRemoveCommentsRanges(t *testing.T) {
	src := []byte("a // c\nb /* d */ c\n")
	tokens := []lexer.Token{
		{Name: "LINE_COMMENT", Start: 2, Stop: 5},
		{Name: "COMMENT", Start: 9, Stop: 15},
	}
	got := RemoveComments(src, tokens, isComment)
	if string(got) != "a \nb  c\n" {
		t.Fatalf("got %q", got)
	}
}

func TestRemoveCommentsIgnoresNonComments(t *testing.T) {
	src := []byte("keep me\n")
	tokens := []lexer.Token{
		{Name: "IDENTIFIER", Start: 0, Stop: 3},
		{Name: "WS", Start: 4, Stop: 4},
	}
	got := RemoveComments(src, tokens, isComment)
	if string(got) != "keep me\n" {
		t.Fatalf("got %q", got)
	}
}

func TestRemoveCommentsOverlap(t *testing.T) {
	src := []byte("abcdef")
	tokens := []lexer.Token{
		{Name: "COMMENT", Start: 1, Stop: 4},
		{Name: "COMMENT", Start: 3, Stop: 5},
	}
	got := RemoveComments(src, tokens, isComment)
	if string(got) != "a" {
		t.Fatalf("got %q", got)
	}
}

func TestRemoveCommentsEmpty(t *testing.T) {
	if got := RemoveComments(nil, nil, isComment); len(got) != 0 {
		t.Fatalf("got %q", got)
	}
}

func TestRemoveCommentsKeepsTrailingNewline(t *testing.T) {
	src := []byte("-- c\nx\n")
	tokens := []lexer.Token{{Name: "LINE_COMMENT", Start: 0, Stop: 4}}
	got := RemoveComments(src, tokens, func(n string) bool { return n == "LINE_COMMENT" })
	if string(got) != "\nx\n" {
		t.Fatalf("got %q", got)
	}
}

func TestStripPreservesShebang(t *testing.T) {
	l, ok := langs.Default().ByName("python")
	if !ok {
		t.Fatal("python not registered")
	}
	got, err := Strip(l, []byte("#!/usr/bin/env python3\n# c\nx = 1\n"))
	if err != nil {
		t.Fatalf("Strip: %v", err)
	}
	if string(got) != "#!/usr/bin/env python3\n\nx = 1\n" {
		t.Fatalf("got %q", got)
	}
}

func TestStripGo(t *testing.T) {
	src := []byte("package main\n\n// remove me\nvar s = \"// keep\"\nfunc main() { /* drop */ }\n")
	want := "package main\n\n\nvar s = \"// keep\"\nfunc main() {  }\n"
	got, err := Strip(golang.New(), src)
	if err != nil {
		t.Fatalf("Strip: %v", err)
	}
	if string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
