package remount

import (
	"io"
	"io/fs"

	"github.com/hack-pad/hackpadfs"
	experimentalsys "github.com/tetratelabs/wazero/experimental/sys"
	"github.com/tetratelabs/wazero/sys"
)

var WazFlags = map[experimentalsys.Oflag]int{
	experimentalsys.O_APPEND: hackpadfs.FlagAppend,
	experimentalsys.O_CREAT:  hackpadfs.FlagCreate,
	experimentalsys.O_RDONLY: hackpadfs.FlagReadOnly,
	experimentalsys.O_WRONLY: hackpadfs.FlagWriteOnly,
	experimentalsys.O_SYNC:   hackpadfs.FlagSync,
	experimentalsys.O_TRUNC:  hackpadfs.FlagTruncate,
}

func GetHackFlag(x experimentalsys.Oflag) int {
	f := 0
	for a, b := range WazFlags {
		if x&a != 0 {
			f |= b
		}
	}
	return f
}

func GetSysFlag(x int) experimentalsys.Oflag {
	var f experimentalsys.Oflag
	for b, a := range WazFlags {
		if x&a != 0 {
			f |= b
		}
	}
	return f
}

func WazAdapt(err error) experimentalsys.Errno {
	if err != nil {
		return experimentalsys.EBADF
	}
	return 0
}

type WazFS struct {
	hackpadfs.FS
}

// Chmod implements sys.FS.
func (WazFS) Chmod(path string, perm fs.FileMode) experimentalsys.Errno {
	panic("unimplemented")
}

// Link implements sys.FS.
func (WazFS) Link(oldPath string, newPath string) experimentalsys.Errno {
	panic("unimplemented")
}

// Lstat implements sys.FS.
func (w WazFS) Lstat(path string) (sys.Stat_t, experimentalsys.Errno) {
	a, err := hackpadfs.Lstat(w.FS, path)
	if err != nil {
		return sys.Stat_t{}, WazAdapt(err)
	}
	return sys.NewStat_t(a), 0
}

// Mkdir implements sys.FS.
func (WazFS) Mkdir(path string, perm fs.FileMode) experimentalsys.Errno {
	panic("unimplemented")
}

// OpenFile implements sys.FS.
func (w *WazFS) OpenFile(path string, flag experimentalsys.Oflag, perm fs.FileMode) (experimentalsys.File, experimentalsys.Errno) {
	wf := &WazFile{
		B:     B{},
		Flags: flag,
		In: struct {
			FS   *WazFS
			Path string
		}{
			FS:   w,
			Path: path,
		},
	}
	err := wf.Reopen()
	if err != nil {
		return nil, WazAdapt(err)
	}
	return wf, 0
}

// Readlink implements sys.FS.
func (WazFS) Readlink(path string) (string, experimentalsys.Errno) {
	panic("unimplemented")
}

// Rename implements sys.FS.
func (WazFS) Rename(from string, to string) experimentalsys.Errno {
	panic("unimplemented")
}

// Rmdir implements sys.FS.
func (WazFS) Rmdir(path string) experimentalsys.Errno {
	panic("unimplemented")
}

// Stat implements sys.FS.
func (w WazFS) Stat(path string) (sys.Stat_t, experimentalsys.Errno) {
	a, err := hackpadfs.Stat(w.FS, path)
	if err != nil {
		return sys.Stat_t{}, WazAdapt(err)
	}
	return sys.NewStat_t(a), 0
}

// Symlink implements sys.FS.
func (WazFS) Symlink(oldPath string, linkName string) experimentalsys.Errno {
	panic("unimplemented")
}

// Unlink implements sys.FS.
func (WazFS) Unlink(path string) experimentalsys.Errno {
	panic("unimplemented")
}

// Utimens implements sys.FS.
func (WazFS) Utimens(path string, atim int64, mtim int64) experimentalsys.Errno {
	panic("unimplemented")
}

var _ experimentalsys.FS = &WazFS{}

type WazFile struct {
	B     B
	Flags experimentalsys.Oflag
	In    struct {
		FS   *WazFS
		Path string
	}
}

func (w *WazFile) Reopen() error {
	var err error
	if w.B.File != nil {
		w.B.Close()
	}
	w.B.File, err = hackpadfs.OpenFile(w.In.FS.FS, w.In.Path, GetHackFlag(w.Flags), 0)
	return err
}

// Close implements sys.File.
func (w WazFile) Close() experimentalsys.Errno {
	return WazAdapt(w.B.Close())
}

// Datasync implements sys.File.
func (w WazFile) Datasync() experimentalsys.Errno {
	return 0
}

// Dev implements sys.File.
func (w WazFile) Dev() (uint64, experimentalsys.Errno) {
	return 0, 0
}

// Ino implements sys.File.
func (WazFile) Ino() (uint64, experimentalsys.Errno) {
	return 0, 0
}

// IsAppend implements sys.File.
func (w WazFile) IsAppend() bool {
	return w.Flags&experimentalsys.O_APPEND != 0
}

// IsDir implements sys.File.
func (w WazFile) IsDir() (bool, experimentalsys.Errno) {
	s, err := w.B.Stat()
	if err != nil {
		return false, WazAdapt(err)
	}
	return s.IsDir(), 0
}

// Pread implements sys.File.
func (w WazFile) Pread(buf []byte, off int64) (n int, errno experimentalsys.Errno) {
	_, err := w.B.Seek(off, io.SeekStart)
	if err != nil {
		return 0, WazAdapt(err)
	}
	a, err := w.B.Read(buf)
	if err != nil {
		return 0, WazAdapt(err)
	}
	return a, 0
}

// Pwrite implements sys.File.
func (w WazFile) Pwrite(buf []byte, off int64) (n int, errno experimentalsys.Errno) {
	_, err := w.B.Seek(off, io.SeekStart)
	if err != nil {
		return 0, WazAdapt(err)
	}
	a, err := w.B.Write(buf)
	if err != nil {
		return 0, WazAdapt(err)
	}
	return a, 0
}

// Read implements sys.File.
func (w WazFile) Read(buf []byte) (n int, errno experimentalsys.Errno) {
	a, err := w.B.Read(buf)
	if err != nil {
		return 0, WazAdapt(err)
	}
	return a, 0
}

// Readdir implements sys.File.
func (w WazFile) Readdir(n int) (dirents []experimentalsys.Dirent, errno experimentalsys.Errno) {
	a, err := w.B.Readdir(n)
	if err != nil {
		return nil, WazAdapt(err)
	}
	b := []experimentalsys.Dirent{}
	for _, c := range a {
		b = append(b, experimentalsys.Dirent{Ino: 0, Name: c.Name(), Type: c.Mode().Type()})
	}
	return b, 0
}

// Seek implements sys.File.
func (w WazFile) Seek(offset int64, whence int) (newOffset int64, errno experimentalsys.Errno) {
	a, err := w.B.Seek(offset, whence)
	if err != nil {
		return 0, WazAdapt(err)
	}
	return a, 0
}

// SetAppend implements sys.File.
func (w *WazFile) SetAppend(enable bool) experimentalsys.Errno {
	if enable {
		w.Flags |= experimentalsys.O_APPEND
	} else {
		w.Flags &= ^experimentalsys.O_APPEND
	}
	w.Flags &= ^experimentalsys.O_CREAT
	return WazAdapt(w.Reopen())
}

// Stat implements sys.File.
func (w WazFile) Stat() (sys.Stat_t, experimentalsys.Errno) {
	a, err := w.B.Stat()
	if err != nil {
		return sys.Stat_t{}, WazAdapt(err)
	}
	return sys.NewStat_t(a), 0
}

// Sync implements sys.File.
func (w WazFile) Sync() experimentalsys.Errno {
	return WazAdapt(w.B.Sync())
}

// Truncate implements sys.File.
func (w WazFile) Truncate(size int64) experimentalsys.Errno {
	return WazAdapt(w.B.Truncate(size))
}

// Utimens implements sys.File.
func (WazFile) Utimens(atim int64, mtim int64) experimentalsys.Errno {
	panic("unimplemented")
}

// Write implements sys.File.
func (w WazFile) Write(buf []byte) (n int, errno experimentalsys.Errno) {
	a, err := w.B.Write(buf)
	if err != nil {
		return 0, WazAdapt(err)
	}
	return a, 0
}

var _ experimentalsys.File = &WazFile{}
