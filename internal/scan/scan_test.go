package scan

import "testing"

func TestScanLine(t *testing.T) {
	s := Syntax{Line: []string{"#"}}
	got := Ranges([]byte("a # c\nb\n"), s)
	if len(got) != 1 || got[0].Start != 2 || got[0].Stop != 5 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanBoundary(t *testing.T) {
	s := Syntax{Line: []string{"#"}, LineBoundary: true}
	if got := Ranges([]byte("a#b\n"), s); len(got) != 0 {
		t.Fatalf("ranges = %+v", got)
	}
	if got := Ranges([]byte("a #b\n"), s); len(got) != 1 {
		t.Fatalf("ranges = %+v", got)
	}
	if got := Ranges([]byte("#b\n"), s); len(got) != 1 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanBlockNested(t *testing.T) {
	s := Syntax{Block: [][2]string{{"/*", "*/"}}, Nest: true}
	got := Ranges([]byte("a /* x /* y */ z */ b\n"), s)
	if len(got) != 1 || got[0].Start != 2 || got[0].Stop != 19 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanBlockLineStart(t *testing.T) {
	s := Syntax{BlockLineStart: [][2]string{{"=begin", "=end"}}}
	got := Ranges([]byte("x\n=begin\ny\n=end\n"), s)
	if len(got) != 1 || string([]byte("x\n=begin\ny\n=end\n")[got[0].Start:got[0].Stop]) != "=begin\ny\n=end" {
		t.Fatalf("ranges = %+v", got)
	}
	if other := Ranges([]byte("x =begin\ny\n=end\n"), s); len(other) != 0 {
		t.Fatalf("ranges = %+v", other)
	}
}

func TestScanStrings(t *testing.T) {
	s := Syntax{Line: []string{"#"}, Strings: []StringRule{{Open: `"`, Close: `"`, Escape: true}}}
	got := Ranges([]byte(`a = "# not" # yes`+"\n"), s)
	if len(got) != 1 || got[0].Start != 12 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanTripleString(t *testing.T) {
	s := Syntax{Line: []string{"#"}, Strings: []StringRule{{Open: `"""`, Close: `"""`, Multiline: true}}}
	got := Ranges([]byte("\"\"\"# not\"\"\"\n# yes\n"), s)
	if len(got) != 1 || got[0].Start != 12 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanVerbatimDoubled(t *testing.T) {
	s := Syntax{Strings: []StringRule{{Open: `@"`, Close: `"`, Doubled: true, Multiline: true}}}
	got := Ranges([]byte("var s = @\"a\"\"b\"; // still string\n"), s)
	if len(got) != 0 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanUnterminated(t *testing.T) {
	s := Syntax{Block: [][2]string{{"/*", "*/"}}}
	got := Ranges([]byte("a /* never ends"), s)
	if len(got) != 1 || got[0].Stop != len("a /* never ends") {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanMarkdown(t *testing.T) {
	s := Syntax{Block: [][2]string{{"<!--", "-->"}}}
	got := Ranges([]byte("x <!-- c --> y\n"), s)
	if len(got) != 1 || got[0].Start != 2 || got[0].Stop != 12 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanEmpty(t *testing.T) {
	if got := Ranges(nil, Syntax{Line: []string{"#"}}); len(got) != 0 {
		t.Fatalf("ranges = %+v", got)
	}
}
