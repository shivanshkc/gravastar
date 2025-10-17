package httputils

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// FileSystemRoot is an http.FileSystem compatible version of the os.Root type.
type FileSystemRoot struct {
	*os.Root
}

// NewFileSystemRoot returns a new instance of FileSystemRoot. Its signature is equivalent to os.OpenRoot.
func NewFileSystemRoot(name string) (*FileSystemRoot, error) {
	root, err := os.OpenRoot(name)
	if err != nil {
		return nil, fmt.Errorf("error in os.OpenRoot call: %w", err)
	}

	return &FileSystemRoot{root}, nil
}

func (f *FileSystemRoot) Open(name string) (http.File, error) {
	name = strings.TrimPrefix(name, "/")
	if name == "" {
		name = "."
	}

	file, err := f.Root.Open(name)
	return http.File(file), err
}
