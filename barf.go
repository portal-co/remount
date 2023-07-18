package remount

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/go-git/go-billy/v5"
	"github.com/hack-pad/hackpadfs"
)

type B struct {
	fs.File
}

func (b B) Readdir(count int) ([]os.FileInfo, error) {
	x, err := hackpadfs.ReadDirFile(b.File, count)
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
func (b B) Readdirnames(count int) ([]string, error) {
	x, err := hackpadfs.ReadDirFile(b.File, count)
	if err != nil {
		return nil, err
	}
	y := []string{}
	for _, z := range x {
		i, err := z.Info()
		if err != nil {
			return nil, err
		}
		y = append(y, i.Name())
	}
	return y, nil
}

func (b B) Name() string {
	return ""
}
func (b B) Write(x []byte) (int, error) {
	return hackpadfs.WriteFile(b.File, x)
}
func (b B) Seek(offset int64, whence int) (int64, error) {
	return hackpadfs.SeekFile(b.File, offset, whence)
}
func (b B) ReadAt(p []byte, off int64) (n int, err error) {
	return hackpadfs.ReadAtFile(b.File, p, off)
}
func (b B) WriteAt(p []byte, off int64) (n int, err error) {
	return hackpadfs.WriteAtFile(b.File, p, off)
}
func (b B) WriteString(s string) (ret int, err error) {
	return b.Write([]byte(s))
}
func (b B) Lock() error {
	return nil
}
func (b B) Sync() error {
	return nil
}
func (b B) Unlock() error {
	return nil
}
func (b B) Truncate(size int64) error {
	return hackpadfs.TruncateFile(b.File, size)
}

type BF struct {
	fs.FS
}

// Create creates the named file with mode 0666 (before umask), truncating
// it if it already exists. If successful, methods on the returned File can
// be used for I/O; the associated file descriptor has mode O_RDWR.
func (b BF) Create(filename string) (billy.File, error) {
	f, err := hackpadfs.Create(b.FS, filename)
	return B{f}, err
}

// Open opens the named file for reading. If successful, methods on the
// returned file can be used for reading; the associated file descriptor has
// mode O_RDONLY.
func (b BF) Open(filename string) (billy.File, error) {
	f, err := b.FS.Open(filename)
	return B{f}, err
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (b BF) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	f, err := hackpadfs.OpenFile(b.FS, filename, flag, perm)
	return B{f}, err
}

// Stat returns a FileInfo describing the named file.
func (b BF) Stat(filename string) (os.FileInfo, error) {
	return hackpadfs.Stat(b.FS, filename)
}

// Rename renames (moves) oldpath to newpath. If newpath already exists and
// is not a directory, Rename replaces it. OS-specific restrictions may
// apply when oldpath and newpath are in different directories.
func (b BF) Rename(oldpath, newpath string) error {
	return hackpadfs.Rename(b.FS, oldpath, newpath)
}

// Remove removes the named file or directory.
func (b BF) Remove(filename string) error {
	return hackpadfs.Remove(b.FS, filename)
}

// Join joins any number of path elements into a single path, adding a
// Separator if necessary. Join calls filepath.Clean on the result; in
// particular, all empty strings are ignored. On Windows, the result is a
// UNC path if and only if the first path element is a UNC path.
func (b BF) Join(elem ...string) string {
	return path.Join(elem...)
}
func (b BF) TempFile(dir, prefix string) (billy.File, error) {
	return nil, fmt.Errorf("not supported (yet): tempfile")
}

// ReadDir reads the directory named by dirname and returns a list of
// directory entries sorted by filename.
func (b BF) ReadDir(path string) ([]os.FileInfo, error) {
	x, err := hackpadfs.ReadDir(b.FS, path)
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
func (b BF) MkdirAll(filename string, perm os.FileMode) error {
	return hackpadfs.MkdirAll(b.FS, filename, perm)
}

// Lstat returns a FileInfo describing the named file. If the file is a
// symbolic link, the returned FileInfo describes the symbolic link. Lstat
// makes no attempt to follow the link.
func (b BF) Lstat(filename string) (os.FileInfo, error) {
	return hackpadfs.Lstat(b.FS, filename)
}

// Symlink creates a symbolic-link from link to target. target may be an
// absolute or relative path, and need not refer to an existing node.
// Parent directories of link are created as necessary.
func (b BF) Symlink(target, link string) error {
	return hackpadfs.Symlink(b.FS, link, target)
}

// Readlink returns the target path of link.
func (b BF) Readlink(link string) (string, error) {
	return "", fmt.Errorf("not supported (yet): readlink")
}

// Chroot returns a new filesystem from the same type where the new root is
// the given path. Files outside of the designated directory tree cannot be
// accessed.
func (b BF) Chroot(path string) (billy.Filesystem, error) {
	s, err := hackpadfs.Sub(b.FS, path)
	return BF{s}, err
}

// Root returns the root path of the filesystem.
func (b BF) Root() string {
	return "/"
}

var _ billy.Filesystem = BF{}

type F struct {
	billy.File
}

func (f F) Stat() (fs.FileInfo, error) {
	return nil, fmt.Errorf("not supported (yet): stat")
}

type FB struct {
	billy.Filesystem
}

// Create creates the named file with mode 0666 (before umask), truncating
// it if it already exists. If successful, methods on the returned File can
// be used for I/O; the associated file descriptor has mode O_RDWR.
func (b FB) Create(filename string) (fs.File, error) {
	x, err := b.Filesystem.Create(filename)
	return F{x}, err
}

// Open opens the named file for reading. If successful, methods on the
// returned file can be used for reading; the associated file descriptor has
// mode O_RDONLY.
func (b FB) Open(filename string) (fs.File, error) {
	x, err := b.Filesystem.Open(filename)
	return F{x}, err
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (b FB) OpenFile(filename string, flag int, perm os.FileMode) (fs.File, error) {
	x, err := b.Filesystem.OpenFile(filename, flag, perm)
	return F{x}, err
}

func (b FB) ReadDir(path string) ([]os.DirEntry, error) {
	x, err := b.Filesystem.ReadDir(path)
	if err != nil {
		return nil, err
	}
	y := []os.DirEntry{}
	for _, z := range x {
		y = append(y, fs.FileInfoToDirEntry(z))
	}
	return y, nil
}

var _ fs.FS = FB{}
