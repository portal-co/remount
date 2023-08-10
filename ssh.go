package remount

import (
	"io/fs"
	"os"

	"github.com/hack-pad/hackpadfs"
	"github.com/pkg/sftp"
)

type Sftp struct {
	*sftp.Client
}

func (s Sftp) Open(filename string) (fs.File, error) {
	f, err := s.Client.Open("/" + filename)
	return f, err
}

func (s Sftp) OpenFile(filename string, flag int, perm os.FileMode) (fs.File, error) {
	f, err := s.Client.OpenFile("/"+filename, flag)
	return f, err
}

var _ hackpadfs.FS = Sftp{}
