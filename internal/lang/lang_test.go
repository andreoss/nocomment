package lang

import (
	"testing"

	"nocomment/internal/lexer"
)

type fakeLanguage struct {
	name string
	exts []string
}

func (f fakeLanguage) Name() string                           { return f.name }
func (f fakeLanguage) Extensions() []string                   { return f.exts }
func (f fakeLanguage) Tokenize([]byte) ([]lexer.Token, error) { return nil, nil }
func (f fakeLanguage) IsComment(string) bool                  { return false }

func TestRegisterAndLookupByName(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeLanguage{name: "go", exts: []string{".go"}})
	l, ok := r.ByName("go")
	if !ok || l.Name() != "go" {
		t.Fatalf("ByName = %v, %v", l, ok)
	}
}

func TestLookupByExtensionNormalizes(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeLanguage{name: "go", exts: []string{"go"}})
	for _, path := range []string{"a.go", "a.GO", "dir/a.go"} {
		if _, ok := r.ByExtension(path); !ok {
			t.Fatalf("ByExtension(%q) not found", path)
		}
	}
}

func TestUnknownLanguage(t *testing.T) {
	r := NewRegistry()
	if _, ok := r.ByName("nope"); ok {
		t.Fatal("unexpected language")
	}
	if _, ok := r.ByExtension("a.zzz"); ok {
		t.Fatal("unexpected extension")
	}
}

func TestLanguages(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeLanguage{name: "go", exts: []string{".go"}})
	r.Register(fakeLanguage{name: "py", exts: []string{".py"}})
	if got := len(r.Languages()); got != 2 {
		t.Fatalf("languages = %d", got)
	}
}

func TestNormalizeExt(t *testing.T) {
	cases := map[string]string{"go": ".go", ".GO": ".go", " .py ": ".py", "": ""}
	for in, want := range cases {
		if got := NormalizeExt(in); got != want {
			t.Fatalf("NormalizeExt(%q) = %q", in, got)
		}
	}
}
