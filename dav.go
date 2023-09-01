package remount

import (
	"context"
	"io/fs"

	"github.com/hack-pad/hackpadfs"
	"golang.org/x/net/webdav"
)

type Dav struct {
	hackpadfs.FS
}

// Mkdir implements webdav.FileSystem.
func (d Dav) Mkdir(ctx context.Context, name string, perm fs.FileMode) error {
	return hackpadfs.Mkdir(d.FS, name, perm)
}

// OpenFile implements webdav.FileSystem.
func (d Dav) OpenFile(ctx context.Context, name string, flag int, perm fs.FileMode) (webdav.File, error) {
	x, err := hackpadfs.OpenFile(d.FS, name, flag, perm)
	return B{x}, err
}

// RemoveAll implements webdav.FileSystem.
func (d Dav) RemoveAll(ctx context.Context, name string) error {
	return hackpadfs.RemoveAll(d.FS, name)
}

// Rename implements webdav.FileSystem.
func (d Dav) Rename(ctx context.Context, oldName string, newName string) error {
	return hackpadfs.Rename(d.FS, oldName, newName)
}

// Stat implements webdav.FileSystem.
func (d Dav) Stat(ctx context.Context, name string) (fs.FileInfo, error) {
	return hackpadfs.Stat(d.FS, name)
}

var _ webdav.FileSystem = Dav{}
