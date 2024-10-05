package cli

import (
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

type memFile struct {
	name string
	data []byte
	mode fs.FileMode
	dir  bool
}

type memFS struct {
	files  map[string]*memFile
	writes int
}

func newMemFS() *memFS {
	return &memFS{files: map[string]*memFile{}}
}

func (m *memFS) put(name string, data []byte, mode fs.FileMode) {
	m.files[name] = &memFile{name: name, data: append([]byte(nil), data...), mode: mode}
}

func (m *memFS) ReadFile(name string) ([]byte, error) {
	f, ok := m.files[name]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}
	if f.dir {
		return nil, &os.PathError{Op: "read", Path: name, Err: fs.ErrInvalid}
	}
	return append([]byte(nil), f.data...), nil
}

func (m *memFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	m.writes++
	m.files[name] = &memFile{name: name, data: append([]byte(nil), data...), mode: perm}
	return nil
}

func (m *memFS) Stat(name string) (fs.FileInfo, error) {
	if f, ok := m.files[name]; ok {
		return memInfo{f}, nil
	}
	prefix := name + "/"
	for fname := range m.files {
		if strings.HasPrefix(fname, prefix) {
			return memInfo{&memFile{name: name, mode: fs.ModeDir | 0o755, dir: true}}, nil
		}
	}
	return nil, &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
}

func (m *memFS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = path.Clean(name)
	if name == "." {
		name = ""
	}
	children := map[string]fs.FileMode{}
	for fname, f := range m.files {
		prefix := name
		if prefix != "" {
			prefix += "/"
		}
		if !strings.HasPrefix(fname, prefix) {
			continue
		}
		rest := fname[len(prefix):]
		if rest == "" {
			continue
		}
		child := rest
		mode := f.mode
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			child = rest[:i]
			mode = fs.ModeDir | 0o755
		}
		if _, ok := children[child]; !ok {
			children[child] = mode
		}
	}
	if len(children) == 0 {
		if f, ok := m.files[name]; ok && f.dir {
			return nil, nil
		}
		return nil, &os.PathError{Op: "readdir", Path: name, Err: os.ErrNotExist}
	}
	out := make([]fs.DirEntry, 0, len(children))
	for n, mode := range children {
		out = append(out, memDirEntry{name: n, mode: mode})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out, nil
}

type memInfo struct{ f *memFile }

func (m memInfo) Name() string       { return m.f.name }
func (m memInfo) Size() int64        { return int64(len(m.f.data)) }
func (m memInfo) Mode() fs.FileMode  { return m.f.mode }
func (m memInfo) ModTime() time.Time { return time.Time{} }
func (m memInfo) IsDir() bool        { return m.f.dir }
func (m memInfo) Sys() any           { return nil }

type memDirEntry struct {
	name string
	mode fs.FileMode
}

func (e memDirEntry) Name() string      { return e.name }
func (e memDirEntry) IsDir() bool       { return e.mode.IsDir() }
func (e memDirEntry) Type() fs.FileMode { return e.mode.Type() }

func (e memDirEntry) Info() (fs.FileInfo, error) {
	return memInfo{&memFile{name: e.name, mode: e.mode, dir: e.mode.IsDir()}}, nil
}
