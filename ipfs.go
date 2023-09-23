package remount

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	gopath "path"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	"github.com/hack-pad/hackpadfs/mount"
	iface "github.com/ipfs/boxo/coreiface"
	"github.com/ipfs/boxo/coreiface/path"
	"github.com/ipfs/boxo/files"
	"go4.org/readerutil"
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
	Rec  func() IF
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

func (i IF) ReadAt(b []byte, off int64) (n int, err error) {
	return readerutil.NewBufferingReaderAt(i.Rec()).ReadAt(b, off)
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
	if x == "" {
		return os.Open("/tmp/portal-ipfs-shim")
	}
	f, err := i.Unixfs().Get(context.Background(), path.New("/ipfs/"+x))
	if err != nil {
		return nil, err
	}
	var r func() IF
	r = func() IF {
		return IF{f, gopath.Base(x), r}
	}
	return r(), nil
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
		return files.NewReaderFile(o), nil
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
func Clone(x fs.FS, dx fs.FS, y, dy string) error {
	o, err := x.Open(y)
	if err != nil {
		return err
	}
	s, err := o.Stat()
	if err != nil {
		o.Close()
		return err
	}
	defer o.Close()
	if !s.IsDir() {
		p, err := hackpadfs.Create(dx, dy)
		if err != nil {
			return err
		}
		defer p.Close()
		_, err = io.Copy(B{p}, o)
		return err
	}
	r, err := hackpadfs.ReadDirFile(o, 0)
	if err != nil {
		return err
	}
	err = hackpadfs.Mkdir(dx, dy, 0777)
	if err != nil {
		return err
	}
	// m := map[string]files.Node{}
	var g errgroup.Group
	for _, s := range r {
		s := s
		g.Go(func() error {
			z := y + "/" + s.Name()
			dz := y + "/" + s.Name()
			return Clone(x, dx, z, dz)
		})
	}
	err = g.Wait()
	return err
}

func Push(i I, x fs.FS, y string) (string, error) {
	n, err := Ipfs(x, y)
	if err != nil {
		return "", err
	}
	u, err := i.Unixfs().Add(context.Background(), n)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(u.String(), "/ipfs/"), nil
}
func NewDir(i I, m map[string]string) (string, error) {
	ft, err := mem.NewFS()
	if err != nil {
		return "", err
	}
	fs, err := mount.NewFS(ft)
	if err != nil {
		return "", err
	}
	for k, v := range m {
		s, err := hackpadfs.Sub(i, v)
		if err != nil {
			return "", err
		}
		fs.AddMount(k, s)
	}
	return Push(i, fs, ".")
}
func Mount(j fs.FS, p string) (func() error, error) {
	f, err := fuse.Mount(p)
	if err != nil {
		return nil, err
	}
	var g errgroup.Group
	g.Go(func() error {
		return fusefs.Serve(f, FuseFS(Dir{j}))
	})
	return func() error {
		err := f.Close()
		if err != nil {
			return err
		}
		g.Wait()
		err = fuse.Unmount(p)
		if err != nil {
			return err
		}
		return nil
	}, nil
}

func MountIpfs(x I, i, p string) (func() error, error) {
	s, err := hackpadfs.Sub(x, i)
	if err != nil {
		return nil, err
	}
	return Mount(s, p)
}

func Patch(i I, x string, f func(*mount.FS) error) (string, error) {
	c, err := hackpadfs.Sub(i, x)
	if err != nil {
		return "", err
	}
	m, err := mem.NewFS()
	if err != nil {
		return "", err
	}
	c = NewCow(c, m)
	d, err := mount.NewFS(c)
	if err != nil {
		return "", err
	}
	err = f(d)
	if err != nil {
		return "", err
	}
	n, err := Ipfs(c, ".")
	if err != nil {
		return "", err
	}
	u, err := i.Unixfs().Add(context.Background(), n)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(u.String(), "/ipfs/"), nil
}
func Meld(i I, x, y string) (string, error) {
	c, err := hackpadfs.Sub(i, x)
	if err != nil {
		return "", err
	}
	d, err := hackpadfs.Sub(i, y)
	if err != nil {
		return "", err
	}
	c = NewCow(c, d)
	n, err := Ipfs(c, ".")
	if err != nil {
		return "", err
	}
	u, err := i.Unixfs().Add(context.Background(), n)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(u.String(), "/ipfs/"), nil
}
