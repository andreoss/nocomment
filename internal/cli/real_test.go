package cli

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func realDir(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "scratch"))
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	dir, err := os.MkdirTemp(root, "realfs-")
	if err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func stripComments(path string, src []byte) ([]byte, error) {
	return bytes.ReplaceAll(src, []byte("// drop\n"), nil), nil
}

type realProcessor struct{}

func (realProcessor) Process(path string, src []byte) ([]byte, error) {
	return stripComments(path, src)
}

func (realProcessor) Supports(path string) bool { return filepath.Ext(path) == ".go" }

func TestRealWriteReplacesContentAndKeepsMode(t *testing.T) {
	dir := realDir(t)
	path := filepath.Join(dir, "a.go")
	if err := os.WriteFile(path, []byte("package a\n// drop\nvar A = 1\n"), 0o640); err != nil {
		t.Fatalf("write: %v", err)
	}
	var out, errBuf bytes.Buffer
	status := Run(Config{Write: true, Paths: []string{path}}, strings.NewReader(""), &out, &errBuf, realProcessor{})
	if status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "package a\nvar A = 1\n" {
		t.Fatalf("content = %q", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o640 {
		t.Fatalf("mode = %v, want -rw-r-----", perm)
	}
}

func TestRealWriteLeavesNoTempFile(t *testing.T) {
	dir := realDir(t)
	path := filepath.Join(dir, "a.go")
	if err := os.WriteFile(path, []byte("package a\n// drop\nvar A = 1\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	var out, errBuf bytes.Buffer
	if status := Run(Config{Write: true, Paths: []string{path}}, strings.NewReader(""), &out, &errBuf, realProcessor{}); status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "a.go" {
		names := []string{}
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("directory holds %v", names)
	}
}

func TestRealWriteRefusesASymlink(t *testing.T) {
	dir := realDir(t)
	target := filepath.Join(dir, "target.go")
	if err := os.WriteFile(target, []byte("package a\n// drop\nvar A = 1\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	link := filepath.Join(dir, "link.go")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	var out, errBuf bytes.Buffer
	status := Run(Config{Write: true, Paths: []string{link}}, strings.NewReader(""), &out, &errBuf, realProcessor{})
	if status != 2 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(got), "// drop") {
		t.Fatalf("target was rewritten through the link: %q", got)
	}
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&fs.ModeSymlink == 0 {
		t.Fatal("link was replaced by a regular file")
	}
}

func TestRealWriteRefusesAReadOnlyFile(t *testing.T) {
	dir := realDir(t)
	path := filepath.Join(dir, "a.go")
	if err := os.WriteFile(path, []byte("package a\n// drop\nvar A = 1\n"), 0o444); err != nil {
		t.Fatalf("write: %v", err)
	}
	var out, errBuf bytes.Buffer
	status := Run(Config{Write: true, Paths: []string{path}}, strings.NewReader(""), &out, &errBuf, realProcessor{})
	if status != 2 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(got), "// drop") {
		t.Fatalf("read-only file was rewritten: %q", got)
	}
}
func TestRealDirectoryWalkWithSelect(t *testing.T) {
	dir := realDir(t)
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for name, body := range map[string]string{
		"a.go":     "package a\n// drop\nvar A = 1\n",
		"sub/c.go": "package c\n// drop\nvar C = 1\n",
		"sub/d.ts": "const d = 1\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	var out, errBuf bytes.Buffer
	cfg := Config{List: true, Select: "*.go", Paths: []string{dir}}
	if status := Run(cfg, strings.NewReader(""), &out, &errBuf, realProcessor{}); status != 0 {
		t.Fatalf("status = %d, stderr = %q", status, errBuf.String())
	}
	listed := strings.Fields(out.String())
	if len(listed) != 2 {
		t.Fatalf("listed %v", listed)
	}
	if filepath.Base(listed[0]) != "a.go" || filepath.Base(listed[1]) != "c.go" {
		t.Fatalf("listed %v", listed)
	}
}

func TestRealInvalidUTF8NamesTheFile(t *testing.T) {
	dir := realDir(t)
	path := filepath.Join(dir, "bad.go")
	if err := os.WriteFile(path, []byte{'p', 'k', 0xff, 0xfe, '\n'}, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	var out, errBuf bytes.Buffer
	processor := failingProcessor{}
	status := Run(Config{Paths: []string{path}}, strings.NewReader(""), &out, &errBuf, processor)
	if status != 2 {
		t.Fatalf("status = %d", status)
	}
	got := errBuf.String()
	if !strings.Contains(got, path) {
		t.Fatalf("error does not name the file: %q", got)
	}
	if !strings.Contains(got, "not valid UTF-8") {
		t.Fatalf("error does not state the reason: %q", got)
	}
}

func TestRealEveryBadFileIsNamed(t *testing.T) {
	dir := realDir(t)
	for _, name := range []string{"one.go", "two.go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte{0xff, 0xfe}, 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	var out, errBuf bytes.Buffer
	cfg := Config{AllErrors: true, Paths: []string{dir}}
	if status := Run(cfg, strings.NewReader(""), &out, &errBuf, failingProcessor{}); status != 2 {
		t.Fatalf("status = %d", status)
	}
	got := errBuf.String()
	for _, name := range []string{"one.go", "two.go"} {
		if !strings.Contains(got, name) {
			t.Fatalf("%s not named: %q", name, got)
		}
	}
}

type failingProcessor struct{}

func (failingProcessor) Process(path string, src []byte) ([]byte, error) {
	if !utf8.Valid(src) {
		return nil, errors.New("input is not valid UTF-8")
	}
	return src, nil
}

func (failingProcessor) Supports(path string) bool { return filepath.Ext(path) == ".go" }
