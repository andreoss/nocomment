package filter

import (
	"testing"

	"nocomment/internal/langs"
)

func stripGo(t *testing.T, src string) string {
	t.Helper()
	l, ok := langs.Default().ByName("go")
	if !ok {
		t.Fatal("go not registered")
	}
	out, err := Strip(l, []byte(src))
	if err != nil {
		t.Fatalf("Strip: %v", err)
	}
	return string(out)
}

func TestWholeLineCommentTakesItsLine(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "line comment alone",
			src:  "package a\n// drop\nvar A = 1\n",
			want: "package a\nvar A = 1\n",
		},
		{
			name: "indented line comment alone",
			src:  "package a\n\tvar A = 1 \n\t// drop\n",
			want: "package a\n\tvar A = 1 \n",
		},
		{
			name: "two comment lines in a row",
			src:  "package a\n// one\n// two\nvar A = 1\n",
			want: "package a\nvar A = 1\n",
		},
		{
			name: "block comment alone on its line",
			src:  "package a\n/* drop */\nvar A = 1\n",
			want: "package a\nvar A = 1\n",
		},
		{
			name: "block comment spanning lines",
			src:  "package a\n/* one\ntwo\nthree */\nvar A = 1\n",
			want: "package a\nvar A = 1\n",
		},
		{
			name: "trailing comment keeps its line",
			src:  "package a\nvar A = 1 // drop\nvar B = 2\n",
			want: "package a\nvar A = 1 \nvar B = 2\n",
		},
		{
			name: "comment before code on the same line keeps the line",
			src:  "package a\n/* drop */ var A = 1\n",
			want: "package a\n var A = 1\n",
		},
		{
			name: "blank line already in the source survives",
			src:  "package a\n\n// drop\n\nvar A = 1\n",
			want: "package a\n\n\nvar A = 1\n",
		},
		{
			name: "comment on the last line without a newline",
			src:  "package a\nvar A = 1\n// drop",
			want: "package a\nvar A = 1\n",
		},
		{
			name: "no comment at all is untouched",
			src:  "package a\n\nvar A = 1\n",
			want: "package a\n\nvar A = 1\n",
		},
		{
			name: "comment text inside a string is untouched",
			src:  "package a\nvar A = \"// keep\"\n",
			want: "package a\nvar A = \"// keep\"\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripGo(t, tc.src); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestLineRemovalNeverJoinsCode(t *testing.T) {
	got := stripGo(t, "package a\nvar A = 1 // drop\nvar B = 2\n")
	if got == "package a\nvar A = 1 var B = 2\n" {
		t.Fatal("lines were joined")
	}
}

func TestShebangLineSurvives(t *testing.T) {
	l, ok := langs.Default().ByName("shell")
	if !ok {
		t.Fatal("shell not registered")
	}
	out, err := Strip(l, []byte("#!/bin/sh\n# drop\necho hi\n"))
	if err != nil {
		t.Fatalf("Strip: %v", err)
	}
	if string(out) != "#!/bin/sh\necho hi\n" {
		t.Fatalf("got %q", out)
	}
}
