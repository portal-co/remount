package remount

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeedleFake/p9"
	"github.com/hack-pad/hackpadfs"
)

// Dir is an implementation of FileSystem that serves from the local
// filesystem. It accepts attachments of either "" or "/", but rejects
// all others.
//
// Note that Dir does not support authentication, simply returning an
// error for any attempt to do so. If authentication is necessary,
// wrap a Dir in an AuthFS instance.
type Dir struct {
	hackpadfs.FS
}

func (d Dir) path(p string) hackpadfs.FS {
	s, _ := hackpadfs.Sub(d.FS, strings.TrimPrefix(p, "/"))
	return s
}

func infoToEntry(fi hackpadfs.FileInfo) p9.DirEntry {
	return p9.DirEntry{
		FileMode:  p9.ModeFromOS(fi.Mode()),
		MTime:     fi.ModTime(),
		Length:    uint64(fi.Size()),
		EntryName: fi.Name(),
	}
}

func infoToEntryD(fi hackpadfs.DirEntry) p9.DirEntry {
	i, _ := fi.Info()
	return infoToEntry(i)
}

// Stat implements Attachment.Stat.
func (d Dir) Stat(p string) (p9.DirEntry, error) {
	fi, err := hackpadfs.Stat(d.FS, p)
	if err != nil {
		return p9.DirEntry{}, err
	}
	return infoToEntry(fi), nil
}

// WriteStat implements Attachment.WriteStat.
func (d Dir) WriteStat(p string, changes p9.StatChanges) error {
	// TODO: Add support for other values.

	// p = d.path(p)
	base := filepath.Dir(p)

	mode, ok := changes.Mode()
	if ok {
		err := hackpadfs.Chmod(d.FS, p, mode.OS())
		if err != nil {
			return err
		}
	}

	atime, ok1 := changes.ATime()
	mtime, ok2 := changes.MTime()
	if ok1 || ok2 {
		err := hackpadfs.Chtimes(d.FS, p, atime, mtime)
		if err != nil {
			return err
		}
	}

	length, ok := changes.Length()
	if ok {
		o, err := d.FS.Open(p)
		if err != nil {
			return err
		}
		err = hackpadfs.TruncateFile(o, int64(length))
		if err != nil {
			return err
		}
	}

	name, ok := changes.Name()
	if ok {
		err := hackpadfs.Rename(d.FS, p, filepath.Join(base, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
	}

	return nil
}

func toOSFlags(mode uint8) (flag int) {
	if mode&p9.OREAD != 0 {
		flag |= os.O_RDONLY
	}
	if mode&p9.OWRITE != 0 {
		flag |= os.O_WRONLY
	}
	if mode&p9.ORDWR != 0 {
		flag |= os.O_RDWR
	}
	if mode&p9.OTRUNC != 0 {
		flag |= os.O_TRUNC
	}
	//if mode&OEXCL != 0 {
	//	flag |= os.O_EXCL
	//}
	//if mode&OAPPEND != 0 {
	//	flag |= os.O_APPEND
	//}

	return flag
}

func fromOSFlags(mode int) (flag uint8) {
	if mode&os.O_RDONLY != 0 {
		flag |= p9.OREAD
	}
	if mode&os.O_WRONLY != 0 {
		flag |= p9.OWRITE
	}
	if mode&os.O_RDWR != 0 {
		flag |= p9.ORDWR
	}
	if mode&os.O_TRUNC != 0 {
		flag |= p9.OTRUNC
	}
	//if mode&OEXCL != 0 {
	//	flag |= os.O_EXCL
	//}
	//if mode&OAPPEND != 0 {
	//	flag |= os.O_APPEND
	//}

	return flag
}

// Auth implements FileSystem.Auth.
func (d Dir) Auth(user, aname string) (p9.File, error) {
	return nil, errors.New("auth not supported")
}

// Attach implements FileSystem.Attach.
func (d Dir) Attach(afile p9.File, user, aname string) (p9.Attachment, error) {
	switch aname {
	case "", "/":
		return d, nil
	}

	return nil, errors.New("unknown attachment")
}

// Open implements Attachment.Open.
func (d Dir) Open(p string, mode uint8) (p9.File, error) {
	flag := toOSFlags(mode)

	file, err := hackpadfs.OpenFile(d.FS, strings.TrimPrefix(p, "/"), flag, 0644)
	return &dirFile{
		File: file,
	}, err
}

// Create implements Attachment.Create.
func (d Dir) Create(p string, perm p9.FileMode, mode uint8) (p9.File, error) {
	// p = d.path(p)

	if perm&p9.ModeDir != 0 {
		err := os.Mkdir(p, os.FileMode(perm.Perm()))
		if err != nil {
			return nil, err
		}
	}

	flag := toOSFlags(mode)

	file, err := hackpadfs.OpenFile(d.FS, p, flag|os.O_CREATE, os.FileMode(perm.Perm()))
	return &dirFile{
		File: file,
	}, err
}

// Remove implements Attachment.Remove.
func (d Dir) Remove(p string) error {
	return hackpadfs.Remove(d.FS, strings.TrimPrefix(p, "/"))
}

type dirFile struct {
	hackpadfs.File
}

func (f *dirFile) ReadAt(p []byte, off int64) (n int, err error) {
	return hackpadfs.ReadAtFile(f.File, p, off)
}

func (f *dirFile) WriteAt(p []byte, off int64) (n int, err error) {
	return hackpadfs.WriteAtFile(f.File, p, off)
}

func (f *dirFile) Readdir() ([]p9.DirEntry, error) {
	// fi, err := f.File.Readdir(-1)
	fi, err := hackpadfs.ReadDirFile(f.File, -1)
	if err != nil {
		return nil, err
	}

	entries := make([]p9.DirEntry, 0, len(fi))
	for _, infoc := range fi {
		info, err := infoc.Info()
		if err != nil {
			return nil, err
		}
		entries = append(entries, infoToEntry(info))
	}
	return entries, nil
}
