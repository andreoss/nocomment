package cli

import (
	"bytes"
	"errors"
	"io/fs"
	"runtime"
	"strings"
	"testing"
	"testing/iotest"
)

type fakeProcessor struct {
	fn       func(path string, src []byte) ([]byte, error)
	supports func(path string) bool
}

func (f fakeProcessor) Process(path string, src []byte) ([]byte, error) {
	if f.fn != nil {
		return f.fn(path, src)
	}
	return append([]byte(nil), src...), nil
}

func (f fakeProcessor) Supports(path string) bool {
	if f.supports != nil {
		return f.supports(path)
	}
	return true
}

func TestParseFlags(t *testing.T) {
	cfg, err := Parse([]string{"-w", "-d", "-l", "-e", "-lang", "go", "a.go", "b.go"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !cfg.Write || !cfg.Diff || !cfg.List || !cfg.AllErrors {
		t.Fatalf("flags = %+v", cfg)
	}
	if cfg.Lang != "go" {
		t.Fatalf("lang = %q", cfg.Lang)
	}
	if len(cfg.Paths) != 2 || cfg.Paths[0] != "a.go" || cfg.Paths[1] != "b.go" {
		t.Fatalf("paths = %v", cfg.Paths)
	}
}

func TestParseUnknownFlag(t *testing.T) {
	if _, err := Parse([]string{"-z"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseHelp(t *testing.T) {
	for _, arg := range []string{"-h", "-help"} {
		cfg, err := Parse([]string{arg})
		if err != nil {
			t.Fatalf("Parse(%q): %v", arg, err)
		}
		if !cfg.Help {
			t.Fatalf("Parse(%q): help not set", arg)
		}
	}
}

func TestHelpWritesUsage(t *testing.T) {
	var out, errBuf bytes.Buffer
	status := Run(Config{Help: true}, strings.NewReader(""), &out, &errBuf, fakeProcessor{})
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(out.String(), "usage:") || !strings.Contains(out.String(), "-w") {
		t.Fatalf("out = %q", out.String())
	}
}

func TestConflictingFlags(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{"write-diff", Config{Write: true, Diff: true, Paths: []string{"a.go"}}},
		{"write-list", Config{Write: true, List: true, Paths: []string{"a.go"}}},
		{"diff-list", Config{Diff: true, List: true, Paths: []string{"a.go"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var errBuf bytes.Buffer
			status := Run(tc.cfg, strings.NewReader(""), &bytes.Buffer{}, &errBuf, fakeProcessor{})
			if status != 2 {
				t.Fatalf("status = %d", status)
			}
			if !strings.Contains(errBuf.String(), "cannot combine") {
				t.Fatalf("stderr = %q", errBuf.String())
			}
		})
	}
}

func TestDefaultPrintsProcessed(t *testing.T) {
	files := newMemFS()
	files.put("a.go", []byte("x // c\n"), 0o644)
	var out, errBuf bytes.Buffer
	processor := fakeProcessor{fn: func(string, []byte) ([]byte, error) { return []byte("x \n"), nil }}
	status := runWith(Config{Paths: []string{"a.go"}}, strings.NewReader(""), &out, &errBuf, processor, files)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	if out.String() != "x \n" {
		t.Fatalf("out = %q", out.String())
	}
	if files.writes != 0 {
		t.Fatalf("writes = %d", files.writes)
	}
}

func TestListOnlyChanged(t *testing.T) {
	files := newMemFS()
	files.put("a.go", []byte("x // c\n"), 0o644)
	files.put("b.go", []byte("y\n"), 0o644)
	processor := fakeProcessor{fn: func(path string, src []byte) ([]byte, error) {
		if path == "a.go" {
			return []byte("x \n"), nil
		}
		return src, nil
	}}
	var out, errBuf bytes.Buffer
	status := runWith(Config{List: true, Paths: []string{"a.go", "b.go"}}, strings.NewReader(""), &out, &errBuf, processor, files)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	if out.String() != "a.go\n" {
		t.Fatalf("out = %q", out.String())
	}
}

func TestWriteOnlyChanged(t *testing.T) {
	files := newMemFS()
	files.put("a.go", []byte("x // c\n"), 0o600)
	files.put("b.go", []byte("y\n"), 0o644)
	processor := fakeProcessor{fn: func(path string, src []byte) ([]byte, error) {
		if path == "a.go" {
			return []byte("x \n"), nil
		}
		return src, nil
	}}
	var out, errBuf bytes.Buffer
	status := runWith(Config{Write: true, Paths: []string{"a.go", "b.go"}}, strings.NewReader(""), &out, &errBuf, processor, files)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	if files.writes != 1 {
		t.Fatalf("writes = %d", files.writes)
	}
	if string(files.files["a.go"].data) != "x \n" {
		t.Fatalf("a.go = %q", files.files["a.go"].data)
	}
	if files.files["a.go"].mode != 0o600 {
		t.Fatalf("mode = %v", files.files["a.go"].mode)
	}
	if string(files.files["b.go"].data) != "y\n" {
		t.Fatalf("b.go = %q", files.files["b.go"].data)
	}
}

func TestDiffShowsChange(t *testing.T) {
	files := newMemFS()
	files.put("a.go", []byte("x // c\n"), 0o644)
	processor := fakeProcessor{fn: func(string, []byte) ([]byte, error) { return []byte("x \n"), nil }}
	var out, errBuf bytes.Buffer
	status := runWith(Config{Diff: true, Paths: []string{"a.go"}}, strings.NewReader(""), &out, &errBuf, processor, files)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	got := out.String()
	for _, want := range []string{"--- a.go", "+++ a.go", "-x // c", "+x "} {
		if !strings.Contains(got, want) {
			t.Fatalf("diff missing %q: %q", want, got)
		}
	}
}

func TestProcessorErrorExit(t *testing.T) {
	files := newMemFS()
	files.put("a.go", []byte("x\n"), 0o644)
	processor := fakeProcessor{fn: func(string, []byte) ([]byte, error) { return nil, errors.New("boom") }}
	var errBuf bytes.Buffer
	status := runWith(Config{Paths: []string{"a.go"}}, strings.NewReader(""), &bytes.Buffer{}, &errBuf, processor, files)
	if status != 2 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(errBuf.String(), "boom") {
		t.Fatalf("stderr = %q", errBuf.String())
	}
}

func TestAllErrorsContinues(t *testing.T) {
	files := newMemFS()
	files.put("a.go", []byte("x\n"), 0o644)
	files.put("b.go", []byte("y\n"), 0o644)
	processor := fakeProcessor{fn: func(path string, src []byte) ([]byte, error) { return nil, errors.New(path + " boom") }}
	var errBuf bytes.Buffer
	status := runWith(Config{AllErrors: true, Paths: []string{"a.go", "b.go"}}, strings.NewReader(""), &bytes.Buffer{}, &errBuf, processor, files)
	if status != 2 {
		t.Fatalf("status = %d", status)
	}
	if n := strings.Count(errBuf.String(), "boom"); n != 2 {
		t.Fatalf("errors = %d, stderr = %q", n, errBuf.String())
	}
}

func TestMissingFile(t *testing.T) {
	var errBuf bytes.Buffer
	status := runWith(Config{Paths: []string{"missing.go"}}, strings.NewReader(""), &bytes.Buffer{}, &errBuf, fakeProcessor{}, newMemFS())
	if status != 2 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(errBuf.String(), "missing.go") {
		t.Fatalf("stderr = %q", errBuf.String())
	}
}

func TestEmptyDirectory(t *testing.T) {
	files := newMemFS()
	files.files["pkg"] = &memFile{name: "pkg", mode: fs.ModeDir, dir: true}
	var out, errBuf bytes.Buffer
	status := runWith(Config{Paths: []string{"pkg"}}, strings.NewReader(""), &out, &errBuf, fakeProcessor{}, files)
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	if out.Len() != 0 {
		t.Fatalf("out = %q", out.String())
	}
}

func TestStdinDefault(t *testing.T) {
	processor := fakeProcessor{fn: func(string, []byte) ([]byte, error) { return []byte("x \n"), nil }}
	var out, errBuf bytes.Buffer
	status := runWith(Config{Lang: "go"}, strings.NewReader("x // c\n"), &out, &errBuf, processor, newMemFS())
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	if out.String() != "x \n" {
		t.Fatalf("out = %q", out.String())
	}
}

func TestStdinList(t *testing.T) {
	processor := fakeProcessor{fn: func(string, []byte) ([]byte, error) { return []byte("x \n"), nil }}
	var out bytes.Buffer
	status := runWith(Config{List: true, Lang: "go"}, strings.NewReader("x // c\n"), &out, &bytes.Buffer{}, processor, newMemFS())
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	if out.String() != "<standard input>\n" {
		t.Fatalf("out = %q", out.String())
	}
}

func TestStdinWriteRejected(t *testing.T) {
	var errBuf bytes.Buffer
	status := runWith(Config{Write: true}, strings.NewReader("x\n"), &bytes.Buffer{}, &errBuf, fakeProcessor{}, newMemFS())
	if status != 2 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(errBuf.String(), "standard input") {
		t.Fatalf("stderr = %q", errBuf.String())
	}
}

func TestUnifiedDiffIdentical(t *testing.T) {
	if got := UnifiedDiff("a.go", []byte("x\n"), []byte("x\n")); len(got) != 0 {
		t.Fatalf("diff = %q", got)
	}
}

func TestStdinDiff(t *testing.T) {
	processor := fakeProcessor{fn: func(string, []byte) ([]byte, error) { return []byte("x \n"), nil }}
	var out bytes.Buffer
	status := runWith(Config{Diff: true, Lang: "go"}, strings.NewReader("x // c\n"), &out, &bytes.Buffer{}, processor, newMemFS())
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(out.String(), "--- <standard input>") {
		t.Fatalf("out = %q", out.String())
	}
}

func TestStdinUnchanged(t *testing.T) {
	var out bytes.Buffer
	status := runWith(Config{List: true, Lang: "go"}, strings.NewReader("x\n"), &out, &bytes.Buffer{}, fakeProcessor{}, newMemFS())
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	if out.Len() != 0 {
		t.Fatalf("out = %q", out.String())
	}
}

func TestStdinProcessorError(t *testing.T) {
	processor := fakeProcessor{fn: func(string, []byte) ([]byte, error) { return nil, errors.New("stdin boom") }}
	var errBuf bytes.Buffer
	status := runWith(Config{Lang: "go"}, strings.NewReader("x\n"), &bytes.Buffer{}, &errBuf, processor, newMemFS())
	if status != 2 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(errBuf.String(), "stdin boom") {
		t.Fatalf("stderr = %q", errBuf.String())
	}
}

func TestStdinReadError(t *testing.T) {
	var errBuf bytes.Buffer
	status := runWith(Config{}, iotest.ErrReader(errors.New("read fail")), &bytes.Buffer{}, &errBuf, fakeProcessor{}, newMemFS())
	if status != 2 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(errBuf.String(), "read fail") {
		t.Fatalf("stderr = %q", errBuf.String())
	}
}

type readFailFS struct{ inner *memFS }

func (r readFailFS) Stat(name string) (fs.FileInfo, error) { return r.inner.Stat(name) }

func (r readFailFS) ReadDir(name string) ([]fs.DirEntry, error) { return r.inner.ReadDir(name) }

func (r readFailFS) ReadFile(string) ([]byte, error) { return nil, errors.New("read fail") }

func (r readFailFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return r.inner.WriteFile(name, data, perm)
}

func TestFileReadError(t *testing.T) {
	mem := newMemFS()
	mem.put("a.go", []byte("x\n"), 0o644)
	var errBuf bytes.Buffer
	status := runWith(Config{Paths: []string{"a.go"}}, strings.NewReader(""), &bytes.Buffer{}, &errBuf, fakeProcessor{}, readFailFS{inner: mem})
	if status != 2 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(errBuf.String(), "read fail") {
		t.Fatalf("stderr = %q", errBuf.String())
	}
}

func TestOSFSMissingPath(t *testing.T) {
	fsys := osFS{}
	if _, err := fsys.Stat("nocomment-absent-file"); err == nil {
		t.Fatal("Stat: expected error")
	}
	if _, err := fsys.ReadFile("nocomment-absent-file"); err == nil {
		t.Fatal("ReadFile: expected error")
	}
	if err := fsys.WriteFile("nocomment-absent-dir/file", nil, 0o644); err == nil {
		t.Fatal("WriteFile: expected error")
	}
}

func TestParseVersion(t *testing.T) {
	cfg, err := Parse([]string{"-version"})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !cfg.Version {
		t.Fatal("version not set")
	}
}

func TestVersionOutput(t *testing.T) {
	var out, errBuf bytes.Buffer
	status := Run(Config{Version: true}, strings.NewReader(""), &out, &errBuf, fakeProcessor{})
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "nocomment") || !strings.Contains(got, runtime.Version()) {
		t.Fatalf("out = %q", got)
	}
}

func TestVersionString(t *testing.T) {
	got := versionString()
	if !strings.HasPrefix(got, "nocomment ") {
		t.Fatalf("got %q", got)
	}
}

func TestHelpMentionsVersion(t *testing.T) {
	if !strings.Contains(usageText, "-version") {
		t.Fatal("usage missing -version")
	}
}

func TestWalkDirectoryFiltersAndOrders(t *testing.T) {
	files := newMemFS()
	files.put("tree/b.go", []byte("b // c\n"), 0o644)
	files.put("tree/a.go", []byte("a // c\n"), 0o644)
	files.put("tree/notes.txt", []byte("x\n"), 0o644)
	files.put("tree/sub/c.go", []byte("c // c\n"), 0o644)
	processor := fakeProcessor{
		fn:       func(path string, src []byte) ([]byte, error) { return append([]byte(path+":"), src...), nil },
		supports: func(path string) bool { return strings.HasSuffix(path, ".go") },
	}
	var out bytes.Buffer
	status := runWith(Config{Paths: []string{"tree"}}, strings.NewReader(""), &out, &bytes.Buffer{}, processor, files)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	want := "tree/a.go:a // c\ntree/b.go:b // c\ntree/sub/c.go:c // c\n"
	if out.String() != want {
		t.Fatalf("got %q want %q", out.String(), want)
	}
}

func TestWalkSkipsVCSAndBuildDirs(t *testing.T) {
	files := newMemFS()
	files.put("tree/.git/config.go", []byte("x\n"), 0o644)
	files.put("tree/vendor/v.go", []byte("x\n"), 0o644)
	files.put("tree/node_modules/n.go", []byte("x\n"), 0o644)
	files.put("tree/.hidden.go", []byte("x\n"), 0o644)
	files.put("tree/keep.go", []byte("k // c\n"), 0o644)
	processor := fakeProcessor{
		fn:       func(path string, src []byte) ([]byte, error) { return append([]byte(path+":"), src...), nil },
		supports: func(path string) bool { return strings.HasSuffix(path, ".go") },
	}
	var out bytes.Buffer
	status := runWith(Config{Paths: []string{"tree"}}, strings.NewReader(""), &out, &bytes.Buffer{}, processor, files)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	if out.String() != "tree/keep.go:k // c\n" {
		t.Fatalf("got %q", out.String())
	}
}

func TestExplicitUnsupportedFileProcessed(t *testing.T) {
	files := newMemFS()
	files.put("x.txt", []byte("x\n"), 0o644)
	called := false
	processor := fakeProcessor{
		fn:       func(path string, src []byte) ([]byte, error) { called = true; return src, nil },
		supports: func(string) bool { return false },
	}
	status := runWith(Config{Paths: []string{"x.txt"}}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, processor, files)
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	if !called {
		t.Fatal("explicit file not processed")
	}
}

func TestStdinShebangInference(t *testing.T) {
	var gotPath string
	processor := fakeProcessor{fn: func(path string, src []byte) ([]byte, error) {
		gotPath = path
		return []byte("x\n"), nil
	}}
	status := runWith(Config{}, strings.NewReader("#!/usr/bin/env python3\n# c\n"), &bytes.Buffer{}, &bytes.Buffer{}, processor, newMemFS())
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	if gotPath != "stdin.py" {
		t.Fatalf("path = %q", gotPath)
	}
}

func TestStdinShebangUnknown(t *testing.T) {
	var errBuf bytes.Buffer
	status := runWith(Config{}, strings.NewReader("#!/usr/bin/perl\n# c\n"), &bytes.Buffer{}, &errBuf, fakeProcessor{}, newMemFS())
	if status != 2 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(errBuf.String(), "requires -lang") {
		t.Fatalf("stderr = %q", errBuf.String())
	}
}

func TestStdinNoShebangRequiresLang(t *testing.T) {
	var errBuf bytes.Buffer
	status := runWith(Config{}, strings.NewReader("x // c\n"), &bytes.Buffer{}, &errBuf, fakeProcessor{}, newMemFS())
	if status != 2 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(errBuf.String(), "requires -lang") {
		t.Fatalf("stderr = %q", errBuf.String())
	}
}

func TestStdinLangOverride(t *testing.T) {
	gotPath := "unset"
	processor := fakeProcessor{fn: func(path string, src []byte) ([]byte, error) {
		gotPath = path
		return src, nil
	}}
	status := runWith(Config{Lang: "go"}, strings.NewReader("package main\n"), &bytes.Buffer{}, &bytes.Buffer{}, processor, newMemFS())
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	if gotPath != "" {
		t.Fatalf("path = %q", gotPath)
	}
}
