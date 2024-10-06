package process

import (
	"strings"
	"testing"

	"nocomment/internal/langs"
)

func newProcessor(language string) Processor {
	return Processor{Registry: langs.Default(), Language: language}
}

func TestProcessByExtension(t *testing.T) {
	src := []byte("package main // c\n")
	got, err := newProcessor("").Process("main.go", src)
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	if string(got) != "package main \n" {
		t.Fatalf("got %q", got)
	}
}

func TestProcessForcedLanguage(t *testing.T) {
	got, err := newProcessor("go").Process("", []byte("// c\n"))
	if err != nil {
		t.Fatalf("Process: %v", err)
	}
	if string(got) != "" {
		t.Fatalf("got %q", got)
	}
}

func TestProcessUnknownLanguage(t *testing.T) {
	_, err := newProcessor("nope").Process("main.go", nil)
	if err == nil || !strings.Contains(err.Error(), "unknown language") {
		t.Fatalf("err = %v", err)
	}
}

func TestProcessStdinRequiresLanguage(t *testing.T) {
	_, err := newProcessor("").Process("", []byte("x\n"))
	if err == nil || !strings.Contains(err.Error(), "standard input") {
		t.Fatalf("err = %v", err)
	}
}

func TestProcessUnsupportedFile(t *testing.T) {
	_, err := newProcessor("").Process("main.zzz", []byte("x\n"))
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("err = %v", err)
	}
}

func TestSupports(t *testing.T) {
	if !newProcessor("").Supports("main.go") {
		t.Fatal("go not supported")
	}
	if newProcessor("").Supports("main.zzz") {
		t.Fatal("zzz supported")
	}
	if !newProcessor("go").Supports("") {
		t.Fatal("forced language not supported")
	}
	if newProcessor("nope").Supports("main.go") {
		t.Fatal("unknown language supported")
	}
}

func TestProcessRejectsNonUTF8(t *testing.T) {
	_, err := newProcessor("go").Process("", []byte{0xff, 0xfe})
	if err == nil || !strings.Contains(err.Error(), "UTF-8") {
		t.Fatalf("err = %v", err)
	}
}

func TestProcessAcceptsBOM(t *testing.T) {
	got, err := newProcessor("go").Process("", []byte("\ufeffpackage main\n"))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !strings.HasPrefix(string(got), "\ufeff") {
		t.Fatalf("BOM lost: %q", got)
	}
}
