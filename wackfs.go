package remount

import "github.com/hack-pad/hackpadfs"

type WackFS func(string) (string, hackpadfs.FS, error)

func (f WackFS) Open(path string) (hackpadfs.File, error) {
	s, t, err := f(path)
	if err != nil {
		return nil, err
	}
	return t.Open(s)
}
func (f WackFS) OpenFile(path string, flag int, perm hackpadfs.FileMode) (hackpadfs.File, error) {
	s, t, err := f(path)
	if err != nil {
		return nil, err
	}
	return hackpadfs.OpenFile(t, s, flag, perm)
}

type Lookup interface {
	LookupFS(string) (string, hackpadfs.FS, error)
}

func L(l Lookup) hackpadfs.FS {
	return WackFS(func(s string) (string, hackpadfs.FS, error) {
		return l.LookupFS(s)
	})
}
