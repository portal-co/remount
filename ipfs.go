package remount

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"time"

	gopath "path"

	"github.com/hack-pad/hackpadfs"
	iface "github.com/ipfs/boxo/coreiface"
	"github.com/ipfs/boxo/coreiface/path"
	"github.com/ipfs/boxo/files"
	"golang.org/x/sync/errgroup"
)

type IN struct {
	name  string
	size  int64
	isDir bool
}

func (i IN) Name() string {
	return i.name
} // base name of the file
func (i IN) Size() int64 {
	return i.size
} // length in bytes for regular files; system-dependent for others
func (i IN) Mode() fs.FileMode {
	var m fs.FileMode
	m = 0777
	if i.isDir {
		m |= fs.ModeDir
	}
	return m
} // file mode bits
func (i IN) ModTime() time.Time {
	return time.Now()
} // modification time
func (i IN) IsDir() bool {
	return i.isDir
} // abbreviation for Mode().IsDir()
func (i IN) Sys() any {
	return nil
} // underlying data source (can return nil)

type IF struct {
	files.Node
	Name string
}

func (i IF) Stat() (os.FileInfo, error) {
	s, err := i.Size()
	if err != nil {
		return nil, err
	}
	_, o := i.Node.(files.Directory)
	return IN{size: s, isDir: o, name: i.Name}, nil
}

func (i IF) Read(x []byte) (int, error) {
	f, ok := i.Node.(files.File)
	if !ok {
		return 0, fmt.Errorf("not supported")
	}
	return f.Read(x)
}

func (i IF) ReadDir(n int) ([]fs.DirEntry, error) {
	f, ok := i.Node.(files.Directory)
	if !ok {
		return nil, fmt.Errorf("not supported")
	}
	x := []fs.DirEntry{}
	it := f.Entries()
	for it.Next() {
		name := it.Name()
		file := it.Node()
		s, err := file.Size()
		if err != nil {
			return nil, err
		}
		_, o := file.(files.Directory)
		x = append(x, fs.FileInfoToDirEntry(IN{name: name, size: s, isDir: o}))
	}
	return x, nil
}

var _ fs.File = IF{}

type I struct {
	iface.CoreAPI
}

func (i I) Open(x string) (fs.File, error) {
	f, err := i.Unixfs().Get(context.Background(), path.New(x))
	if err != nil {
		return nil, err
	}
	return IF{f, gopath.Base(x)}, nil
}

var _ fs.FS = I{}

type N struct {
	fs.File
	FS   fs.FS
	Name string
}

func (n N) Size() (int64, error) {
	s, err := hackpadfs.Stat(n.FS, n.Name)
	if err != nil {
		return 0, err
	}
	return s.Size(), nil
}

var _ files.Node = N{}

func Ipfs(x fs.FS, y string) (files.Node, error) {
	o, err := x.Open(y)
	if err != nil {
		return nil, err
	}
	s, err := o.Stat()
	if err != nil {
		o.Close()
		return nil, err
	}
	if !s.IsDir() {
		return N{o, x, y}, nil
	}
	defer o.Close()
	r, err := hackpadfs.ReadDirFile(o, 0)
	if err != nil {
		return nil, err
	}
	m := map[string]files.Node{}
	var g errgroup.Group
	for _, s := range r {
		s := s
		g.Go(func() error {
			z := y + "/" + s.Name()
			n, err := Ipfs(x, z)
			if err != nil {
				return err
			}
			m[s.Name()] = n
			return nil
		})
	}
	err = g.Wait()
	if err != nil {
		return nil, err
	}
	return files.NewMapDirectory(m), nil
}
