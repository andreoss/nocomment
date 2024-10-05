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

func TestScanRegex(t *testing.T) {
	s := Syntax{Line: []string{"//"}, Block: [][2]string{{"/*", "*/"}}, Regex: true}
	src := []byte("const re = /a\\/\\//; // drop\n")
	got := Ranges(src, s)
	if len(got) != 1 || string(src[got[0].Start:got[0].Stop]) != "// drop" {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanRegexDivision(t *testing.T) {
	s := Syntax{Line: []string{"//"}, Regex: true}
	got := Ranges([]byte("a / b // c\n"), s)
	if len(got) != 1 || got[0].Start != 6 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanRegexAfterKeyword(t *testing.T) {
	s := Syntax{Line: []string{"//"}, Regex: true}
	got := Ranges([]byte("return /a/ // c\n"), s)
	if len(got) != 1 || got[0].Start != 11 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanPercent(t *testing.T) {
	s := Syntax{Line: []string{"#"}, Percent: true}
	got := Ranges([]byte("s = %q{# x}\n# c\n"), s)
	if len(got) != 1 || got[0].Start != 12 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanPercentModulo(t *testing.T) {
	s := Syntax{Line: []string{"#"}, Percent: true}
	got := Ranges([]byte("a % b # c\n"), s)
	if len(got) != 1 || got[0].Start != 6 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanHereDoc(t *testing.T) {
	s := Syntax{Line: []string{"#"}, LineBoundary: true, HereDocs: "shell"}
	got := Ranges([]byte("cat <<EOF\n# x\nEOF\n# c\n"), s)
	if len(got) != 1 || got[0].Start != 18 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanHereDocShift(t *testing.T) {
	s := Syntax{Line: []string{"#"}, HereDocs: "shell"}
	got := Ranges([]byte("x=$((1 << 2)) # c\n"), s)
	if len(got) != 1 || got[0].Start != 14 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanPHPHereDoc(t *testing.T) {
	s := Syntax{Line: []string{"//"}, HereDocs: "php"}
	got := Ranges([]byte("<?php\n$s = <<<EOT\n# x\nEOT;\n// c\n"), s)
	if len(got) != 1 || got[0].Start != 27 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanYAMLBlock(t *testing.T) {
	s := Syntax{Line: []string{"#"}, LineBoundary: true, YAMLBlock: true}
	got := Ranges([]byte("key: |\n  # x\n# c\n"), s)
	if len(got) != 1 || got[0].Start != 13 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanPercentDelimiters(t *testing.T) {
	s := Syntax{Line: []string{"#"}, Percent: true}
	for _, src := range []string{"%w[a b]\n# c\n", "%i{a b}\n# c\n", "%r<a/b>\n# c\n", "%s(x)\n# c\n"} {
		if got := Ranges([]byte(src), s); len(got) != 1 {
			t.Fatalf("%q ranges = %+v", src, got)
		}
	}
}

func TestScanPercentUnterminated(t *testing.T) {
	s := Syntax{Line: []string{"#"}, Percent: true}
	if got := Ranges([]byte("%q{abc"), s); len(got) != 0 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanRegexClassAndUnterminated(t *testing.T) {
	s := Syntax{Line: []string{"//"}, Regex: true}
	got := Ranges([]byte("x = /[a/]/\n// c\n"), s)
	if len(got) != 1 || got[0].Start != 11 {
		t.Fatalf("ranges = %+v", got)
	}
	if got := Ranges([]byte("x = /abc\n"), s); len(got) != 0 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanHereDocVariants(t *testing.T) {
	s := Syntax{Line: []string{"#"}, HereDocs: "shell"}
	for _, src := range []string{
		"cat <<-EOF\n\t# x\n\tEOF\n# c\n",
		"cat <<'EOF'\n# x\nEOF\n# c\n",
		"cat <<\"EOF\"\n# x\nEOF\n# c\n",
	} {
		if got := Ranges([]byte(src), s); len(got) != 1 {
			t.Fatalf("%q ranges = %+v", src, got)
		}
	}
}

func TestScanHereDocNoDelimiter(t *testing.T) {
	s := Syntax{Line: []string{"#"}, HereDocs: "shell"}
	if got := Ranges([]byte("echo << # c\n"), s); len(got) != 1 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanHereDocUnterminated(t *testing.T) {
	s := Syntax{Line: []string{"#"}, HereDocs: "shell"}
	if got := Ranges([]byte("cat <<EOF\n# x\n"), s); len(got) != 0 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanPHPHereDocVariants(t *testing.T) {
	s := Syntax{Line: []string{"//"}, HereDocs: "php"}
	for _, src := range []string{
		"<?php\n$s = <<<'EOT'\n# x\nEOT;\n// c\n",
		"<?php\n$s = <<<EOT\n# x\n    EOT;\n// c\n",
	} {
		if got := Ranges([]byte(src), s); len(got) != 1 {
			t.Fatalf("%q ranges = %+v", src, got)
		}
	}
}

func TestScanPHPHereDocNoDelimiter(t *testing.T) {
	s := Syntax{Line: []string{"//"}, HereDocs: "php"}
	if got := Ranges([]byte("<?php <<< // c\n"), s); len(got) != 1 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanYAMLBlockVariants(t *testing.T) {
	s := Syntax{Line: []string{"#"}, YAMLBlock: true}
	if got := Ranges([]byte("key: >-\n  # x\n\n  y\n# c\n"), s); len(got) != 1 {
		t.Fatalf("ranges = %+v", got)
	}
	if got := Ranges([]byte("key: | value # c\n"), s); len(got) != 1 {
		t.Fatalf("ranges = %+v", got)
	}
}

func TestScanStringUnterminated(t *testing.T) {
	s := Syntax{Line: []string{"#"}, Strings: []StringRule{{Open: `"`, Close: `"`, Escape: true}}}
	got := Ranges([]byte("a = \"unterminated\n# c\n"), s)
	if len(got) != 1 || got[0].Start != 18 {
		t.Fatalf("ranges = %+v", got)
	}
}
