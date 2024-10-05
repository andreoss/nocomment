package cli

import (
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type memFile struct {
	name string
	data []byte
	mode fs.FileMode
	dir  bool
	link string
}

type memFS struct {
	files     map[string]*memFile
	writes    int
	creates   int
	renames   int
	removes   int
	writeErr  error
	chmodErr  error
	renameErr error
}

func newMemFS() *memFS {
	return &memFS{files: map[string]*memFile{}}
}

func (m *memFS) put(name string, data []byte, mode fs.FileMode) {
	m.files[name] = &memFile{name: name, data: append([]byte(nil), data...), mode: mode}
}

func (m *memFS) putLink(name, target string) {
	m.files[name] = &memFile{name: name, link: target, mode: fs.ModeSymlink | 0o777}
}

func (m *memFS) ReadFile(name string) ([]byte, error) {
	f, ok := m.files[name]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}
	if f.link != "" {
		return m.ReadFile(f.link)
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
		if f.link != "" {
			if target, ok := m.files[f.link]; ok {
				return memInfo{target}, nil
			}
		}
		return memInfo{f}, nil
	}
	if info, ok := m.synthDir(name); ok {
		return info, nil
	}
	return nil, &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
}

func (m *memFS) Lstat(name string) (fs.FileInfo, error) {
	if f, ok := m.files[name]; ok {
		return memInfo{f}, nil
	}
	if info, ok := m.synthDir(name); ok {
		return info, nil
	}
	return nil, &os.PathError{Op: "lstat", Path: name, Err: os.ErrNotExist}
}

func (m *memFS) synthDir(name string) (fs.FileInfo, bool) {
	prefix := name + "/"
	for fname := range m.files {
		if strings.HasPrefix(fname, prefix) {
			return memInfo{&memFile{name: name, mode: fs.ModeDir | 0o755, dir: true}}, true
		}
	}
	return nil, false
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

func (m *memFS) Create(name string) (io.WriteCloser, error) {
	m.creates++
	m.files[name] = &memFile{name: name, mode: 0o666}
	return &memWriter{m: m, name: name}, nil
}

func (m *memFS) CreateTemp(dir, pattern string) (TempFile, error) {
	m.creates++
	name := filepath.Join(dir, strings.Replace(pattern, "*", strconv.Itoa(m.creates), 1))
	m.files[name] = &memFile{name: name, mode: 0o600}
	return &memWriter{m: m, name: name}, nil
}

func (m *memFS) Chmod(name string, mode fs.FileMode) error {
	if m.chmodErr != nil {
		return m.chmodErr
	}
	f, ok := m.files[name]
	if !ok {
		return &os.PathError{Op: "chmod", Path: name, Err: os.ErrNotExist}
	}
	f.mode = mode
	return nil
}

func (m *memFS) Rename(oldpath, newpath string) error {
	if m.renameErr != nil {
		return m.renameErr
	}
	f, ok := m.files[oldpath]
	if !ok {
		return &os.PathError{Op: "rename", Path: oldpath, Err: os.ErrNotExist}
	}
	delete(m.files, oldpath)
	f.name = newpath
	m.files[newpath] = f
	m.renames++
	return nil
}

func (m *memFS) Remove(name string) error {
	if _, ok := m.files[name]; !ok {
		return &os.PathError{Op: "remove", Path: name, Err: os.ErrNotExist}
	}
	delete(m.files, name)
	m.removes++
	return nil
}

type memWriter struct {
	m    *memFS
	name string
}

func (w *memWriter) Write(p []byte) (int, error) {
	if w.m.writeErr != nil {
		return 0, w.m.writeErr
	}
	f := w.m.files[w.name]
	if f == nil {
		return 0, os.ErrNotExist
	}
	f.data = append(f.data, p...)
	return len(p), nil
}

func (w *memWriter) Close() error { return nil }

func (w *memWriter) Name() string { return w.name }

type memInfo struct{ f *memFile }

func (m memInfo) Name() string { return m.f.name }

func (m memInfo) Size() int64 { return int64(len(m.f.data)) }

func (m memInfo) Mode() fs.FileMode {
	if m.f.link != "" {
		return fs.ModeSymlink | 0o777
	}
	return m.f.mode
}

func (m memInfo) ModTime() time.Time { return time.Time{} }

func (m memInfo) IsDir() bool { return m.f.dir }

func (m memInfo) Sys() any { return nil }

type memDirEntry struct {
	name string
	mode fs.FileMode
}

func (e memDirEntry) Name() string { return e.name }

func (e memDirEntry) IsDir() bool { return e.mode.IsDir() }

func (e memDirEntry) Type() fs.FileMode { return e.mode.Type() }

func (e memDirEntry) Info() (fs.FileInfo, error) {
	return memInfo{&memFile{name: e.name, mode: e.mode, dir: e.mode.IsDir()}}, nil
}
