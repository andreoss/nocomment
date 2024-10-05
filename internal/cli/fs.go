package cli

import (
	"io"
	"io/fs"
	"os"
)

type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Stat(name string) (fs.FileInfo, error)
	Lstat(name string) (fs.FileInfo, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	Create(name string) (io.WriteCloser, error)
	CreateTemp(dir, pattern string) (TempFile, error)
	Chmod(name string, mode fs.FileMode) error
	Rename(oldpath, newpath string) error
	Remove(name string) error
}

type TempFile interface {
	io.WriteCloser
	Name() string
}

type osFS struct{}

func (osFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }

func (osFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (osFS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }

func (osFS) Lstat(name string) (fs.FileInfo, error) { return os.Lstat(name) }

func (osFS) ReadDir(name string) ([]fs.DirEntry, error) { return os.ReadDir(name) }

func (osFS) Create(name string) (io.WriteCloser, error) { return os.Create(name) }

func (osFS) CreateTemp(dir, pattern string) (TempFile, error) { return os.CreateTemp(dir, pattern) }

func (osFS) Chmod(name string, mode fs.FileMode) error { return os.Chmod(name, mode) }

func (osFS) Rename(oldpath, newpath string) error { return os.Rename(oldpath, newpath) }

func (osFS) Remove(name string) error { return os.Remove(name) }
