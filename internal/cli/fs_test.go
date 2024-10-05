package cli

import (
	"io/fs"
	"os"
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
	f, ok := m.files[name]
	if !ok {
		return nil, &os.PathError{Op: "stat", Path: name, Err: os.ErrNotExist}
	}
	return memInfo{f}, nil
}

type memInfo struct{ f *memFile }

func (m memInfo) Name() string       { return m.f.name }
func (m memInfo) Size() int64        { return int64(len(m.f.data)) }
func (m memInfo) Mode() fs.FileMode  { return m.f.mode }
func (m memInfo) ModTime() time.Time { return time.Time{} }
func (m memInfo) IsDir() bool        { return m.f.dir }
func (m memInfo) Sys() any           { return nil }
