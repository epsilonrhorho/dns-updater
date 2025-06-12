package storage

import (
	"os"
	"strings"
)

// Interface defines the behavior for storing and retrieving IP addresses.
type Interface interface {
	ReadLastIP() (string, error)
	WriteIP(ip string) error
}

// FileStorage implements Interface using the file system.
type FileStorage struct {
	path string
}

// NewFileStorage creates a new FileStorage instance.
func NewFileStorage(path string) *FileStorage {
	return &FileStorage{path: path}
}

// ReadLastIP reads the last stored IP address from the file.
func (fs *FileStorage) ReadLastIP() (string, error) {
	data, err := os.ReadFile(fs.path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteIP writes the IP address to the file.
func (fs *FileStorage) WriteIP(ip string) error {
	return os.WriteFile(fs.path, []byte(ip), 0600)
}