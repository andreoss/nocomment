package langs

import "testing"

func TestDefault(t *testing.T) {
	r := Default()
	if _, ok := r.ByName("go"); !ok {
		t.Fatal("go not registered")
	}
	if _, ok := r.ByExtension("main.go"); !ok {
		t.Fatal(".go not registered")
	}
}

func TestTokenizers(t *testing.T) {
	r := Default()
	cases := []struct {
		language string
		source   string
	}{
		{"go", "package main // c\n"},
		{"java", "class A {} // c\n"},
		{"c", "int x; // c\n"},
		{"cpp", "int x; // c\n"},
		{"typescript", "const s = 1 // c\n"},
		{"javascript", "const s = 1 // c\n"},
		{"kotlin", "val s = 1 // c\n"},
		{"rust", "let s = 1; // c\n"},
		{"scala", "val s = 1 // c\n"},
		{"sqlite", "select 1; -- c\n"},
		{"sql", "select 1; -- c\n"},
		{"postgresql", "select 1; -- c\n"},
		{"toml", "k = 1 # c\n"},
		{"html", "<p>x</p><!-- c -->\n"},
	}
	for _, tc := range cases {
		t.Run(tc.language, func(t *testing.T) {
			l, ok := r.ByName(tc.language)
			if !ok {
				t.Fatalf("%s not registered", tc.language)
			}
			tokens, err := l.Tokenize([]byte(tc.source))
			if err != nil {
				t.Fatalf("Tokenize: %v", err)
			}
			if len(tokens) == 0 {
				t.Fatal("no tokens")
			}
			found := false
			for _, tok := range tokens {
				if l.IsComment(tok.Name) {
					found = true
				}
			}
			if !found {
				t.Fatalf("no comment token in %+v", tokens)
			}
		})
	}
}

func TestExtensions(t *testing.T) {
	r := Default()
	for ext, want := range map[string]string{
		".go": "go", ".java": "java", ".c": "c", ".h": "c", ".cpp": "cpp",
		".js": "javascript", ".mjs": "javascript", ".ts": "typescript",
		".kt": "kotlin", ".rs": "rust", ".scala": "scala", ".sqlite": "sqlite",
		".sql": "sql", ".pgsql": "postgresql", ".toml": "toml",
		".html": "html", ".htm": "html",
	} {
		l, ok := r.ByExtension("file" + ext)
		if !ok {
			t.Fatalf("%s not registered", ext)
		}
		if l.Name() != want {
			t.Fatalf("%s -> %s, want %s", ext, l.Name(), want)
		}
	}
}
