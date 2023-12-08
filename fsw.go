package remount

import (
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/DeedleFake/p9"
	"github.com/hack-pad/hackpadfs"
)

type PathF struct {
	p9.Attachment
	Path string
}

type FSF struct {
	p9.File
	Path   PathF
	Offset *int64
}

type FSS struct {
	p9.DirEntry
}

func (f FSF) Read(p []byte) (int, error) {
	z, err := f.ReadAt(p, *f.Offset)
	*f.Offset += int64(z)
	return z, err
}

func (f FSF) Write(p []byte) (int, error) {
	z, err := f.WriteAt(p, *f.Offset)
	*f.Offset += int64(z)
	return z, err
}

func (f FSF) ReadDir(n int) ([]hackpadfs.DirEntry, error) {
	d, err := f.File.Readdir()
	if err != nil {
		return nil, err
	}
	e := make([]hackpadfs.DirEntry, len(d))
	for i, v := range d {
		e[i] = fs.FileInfoToDirEntry(FSS{v})
	}
	return e, nil
}

func (f FSF) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekCurrent {
		*f.Offset += offset
		return *f.Offset, nil
	}
	if whence == io.SeekStart {
		*f.Offset = offset
		return *f.Offset, nil
	}
	return *f.Offset, fmt.Errorf("not supported")
}

func (f FSF) Stat() (hackpadfs.FileInfo, error) {
	z, err := f.Path.Attachment.Stat(f.Path.Path)
	return FSS{z}, err
}

type FSW struct {
	p9.Attachment
}

func (f FSW) Open(x string) (hackpadfs.File, error) {
	y, err := f.Attachment.Open(x, p9.OREAD)
	var n int64
	return FSF{File: y, Offset: &n, Path: PathF{Attachment: f.Attachment, Path: x}}, err
}
func (f FSW) OpenFile(x string, flag int, perm hackpadfs.FileMode) (hackpadfs.File, error) {
	if flag%os.O_CREATE == 0 {
		y, err := f.Attachment.Open(x, fromOSFlags(flag))
		var n int64
		return FSF{File: y, Offset: &n, Path: PathF{Attachment: f.Attachment, Path: x}}, err
	} else {
		y, err := f.Attachment.Create(x, p9.ModeFromOS(perm), fromOSFlags(flag))
		var n int64
		return FSF{File: y, Offset: &n, Path: PathF{Attachment: f.Attachment, Path: x}}, err
	}
}

type FSP struct {
	*p9.Remote
}

func (f FSP) Create(a string, b p9.FileMode, c uint8) (p9.File, error) {
	return f.Remote.Create(a, b, c)
}
func (f FSP) Open(a string, b uint8) (p9.File, error) {
	return f.Remote.Open(a, b)
}
func (f FSP) WriteStat(path string, changes p9.StatChanges) error {
	return fmt.Errorf("not supported")
}

var _ p9.Attachment = FSP{}
