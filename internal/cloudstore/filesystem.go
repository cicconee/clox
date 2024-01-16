package cloudstore

import (
	"io/fs"
	"os"
)

// OSFileSystem is a wrapper around the os package file system functions.
type OSFileSystem struct{}

// Set calls the os.Stat function.
//
// Stat returns a FileInfo describing the named file. If there is an error, it
// will be of type *PathError.
func (fs *OSFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// Mkdir calls the os.Mkdir function.
//
// Mkdir creates a new directory with the specified name and permission bits
// (before umask). If there is an error, it will be of type *PathError.
func (fs *OSFileSystem) Mkdir(name string, perm fs.FileMode) error {
	return os.Mkdir(name, perm)
}

// Create calls the os.OpenFile function.
//
// The file is created on the file system with the specified name and permissions.
// Name is the full path to the file. The file is created with the os.O_WRONLY,
// os.O_CREATE, and os.O_TRUNC flags. These flags open the file as write only,
// creates the file if it does not exist, and if the file exists, truncate it to
// zero.
func (fs *OSFileSystem) Create(name string, perm fs.FileMode) (*os.File, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
}

// RemoveAll calls the os.RemoveAll function.
//
// RemoveAll removes path and any children it contains. It removes everything
// it can but returns the first error it encounters. If the path does not exist,
// RemoveAll returns nil (no error). If there is an error, it will be of type
// *PathError.
func (fs *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// IsNotExist calls the os.IsNotExist function.
//
// IsNotExist returns a boolean indicating whether the error is known to
// report that a file or directory does not exist. It is satisfied by
// ErrNotExist as well as some syscall errors.
//
// This function predates errors.Is. It only supports errors returned by
// the os package. New code should use errors.Is(err, fs.ErrNotExist).
func (fs *OSFileSystem) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}
