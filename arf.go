package remount

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/hack-pad/hackpadfs"
	"github.com/spf13/afero"
)

func ar(x string) string {
	return strings.TrimPrefix(path.Clean(x), "/")
}

type AF struct {
	fs.FS
}

// Create creates the named file with mode 0666 (before umask), truncating
// it if it already exists. If successful, methods on the returned File can
// be used for I/O; the associated file descriptor has mode O_RDWR.
func (b AF) Create(filename string) (afero.File, error) {
	f, err := hackpadfs.Create(b.FS, ar(filename))
	return B{f}, err
}

// Open opens the named file for reading. If successful, methods on the
// returned file can be used for reading; the associated file descriptor has
// mode O_RDONLY.
func (b AF) Open(filename string) (afero.File, error) {
	f, err := b.FS.Open(ar(filename))
	return B{f}, err
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (b AF) OpenFile(filename string, flag int, perm os.FileMode) (afero.File, error) {
	f, err := hackpadfs.OpenFile(b.FS, ar(filename), flag, perm)
	return B{f}, err
}

// Stat returns a FileInfo describing the named file.
func (b AF) Stat(filename string) (os.FileInfo, error) {
	return hackpadfs.Stat(b.FS, ar(filename))
}

// Rename renames (moves) oldpath to newpath. If newpath already exists and
// is not a directory, Rename replaces it. OS-specific restrictions may
// apply when oldpath and newpath are in different directories.
func (b AF) Rename(oldpath, newpath string) error {
	return hackpadfs.Rename(b.FS, ar(oldpath), ar(newpath))
}

// Remove removes the named file or directory.
func (b AF) Remove(filename string) error {
	return hackpadfs.Remove(b.FS, ar(filename))
}

func (b AF) RemoveAll(filename string) error {
	return hackpadfs.RemoveAll(b.FS, ar(filename))
}

// Join joins any number of path elements into a single path, adding a
// Separator if necessary. Join calls filepath.Clean on the result; in
// particular, all empty strings are ignored. On Windows, the result is a
// UNC path if and only if the first path element is a UNC path.
func (b AF) Join(elem ...string) string {
	return path.Join(elem...)
}
func (b AF) TempFile(dir, prefix string) (billy.File, error) {
	return nil, fmt.Errorf("not supported (yet): tempfile")
}

// ReadDir reads the directory named by dirname and returns a list of
// directory entries sorted by filename.
func (b AF) ReadDir(path string) ([]os.FileInfo, error) {
	x, err := hackpadfs.ReadDir(b.FS, ar(path))
	if err != nil {
		return nil, err
	}
	y := []os.FileInfo{}
	for _, z := range x {
		i, err := z.Info()
		if err != nil {
			return nil, err
		}
		y = append(y, i)
	}
	return y, nil
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error. The permission bits
// perm are used for all directories that MkdirAll creates. If path is/
// already a directory, MkdirAll does nothing and returns nil.
func (b AF) MkdirAll(filename string, perm os.FileMode) error {
	return hackpadfs.MkdirAll(b.FS, ar(filename), perm)
}

func (b AF) Mkdir(filename string, perm os.FileMode) error {
	return hackpadfs.Mkdir(b.FS, ar(filename), perm)
}

// Lstat returns a FileInfo describing the named file. If the file is a
// symbolic link, the returned FileInfo describes the symbolic link. Lstat
// makes no attempt to follow the link.
func (b AF) Lstat(filename string) (os.FileInfo, error) {
	return hackpadfs.Lstat(b.FS, ar(filename))
}

// Symlink creates a symbolic-link from link to target. target may be an
// absolute or relative path, and need not refer to an existing node.
// Parent directories of link are created as necessary.
func (b AF) Symlink(target, link string) error {
	return hackpadfs.Symlink(b.FS, ar(link), ar(target))
}

// Readlink returns the target path of link.
func (b AF) Readlink(link string) (string, error) {
	return "", fmt.Errorf("not supported (yet): readlink")
}

// Root returns the root path of the filesystem.
func (b AF) Root() string {
	return "/"
}

// Chmod changes the mode of the named file to mode.
func (b AF) Chmod(name string, mode os.FileMode) error {
	return nil
}

// Chown changes the uid and gid of the named file.
func (b AF) Chown(name string, uid, gid int) error {
	return nil
}

// Chtimes changes the access and modification times of the named file
func (b AF) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return nil
}

func (b AF) Name() string {
	return ""
}

var _ afero.Fs = AF{}

type A struct {
	afero.File
}

type FA struct {
	afero.Afero
}

// Create creates the named file with mode 0666 (before umask), truncating
// it if it already exists. If successful, methods on the returned File can
// be used for I/O; the associated file descriptor has mode O_RDWR.
func (b FA) Create(filename string) (fs.File, error) {
	x, err := b.Afero.Create("/" + filename)
	return A{x}, err
}

// Open opens the named file for reading. If successful, methods on the
// returned file can be used for reading; the associated file descriptor has
// mode O_RDONLY.
func (b FA) Open(filename string) (fs.File, error) {
	x, err := b.Afero.Open("/" + filename)
	return A{x}, err
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (b FA) OpenFile(filename string, flag int, perm os.FileMode) (fs.File, error) {
	x, err := b.Afero.OpenFile("/"+filename, flag, perm)
	return A{x}, err
}

func (b FA) ReadDir(path string) ([]os.DirEntry, error) {
	x, err := b.Afero.ReadDir("/" + path)
	if err != nil {
		return nil, err
	}
	y := []os.DirEntry{}
	for _, z := range x {
		y = append(y, fs.FileInfoToDirEntry(z))
	}
	return y, nil
}

func NewAferoShim(x afero.Fs) FA {
	return FA{afero.Afero{x}}
}

func NewCow(over fs.FS, layer fs.FS) fs.FS {
	return NewAferoShim(afero.NewCopyOnWriteFs(AF{over}, AF{layer}))
}

var _ fs.FS = FA{}
