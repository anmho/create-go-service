package generator

import (
	"io/fs"
	"os"
)

// FileSystem abstracts file system operations for testability
type FileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	WriteFile(name string, data []byte, perm os.FileMode) error
	ReadFile(name string) ([]byte, error)
	Stat(name string) (fs.FileInfo, error)
	RemoveAll(path string) error
}

// OSFileSystem implements FileSystem using the real OS
type OSFileSystem struct{}

func (f *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (f *OSFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (f *OSFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (f *OSFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (f *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

