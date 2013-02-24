package ws

import (
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"sync"
)

const (
	FlagDir uint64 = 1 << iota
	FlagLogical
	FlagMount
	FlagIgnore
)

type Id uint32

func NewId(path string) Id {
	h := fnv.New32()
	h.Write([]byte(path))
	return Id(h.Sum32())
}

type Dir struct {
	Path     string
	Children []*Res
}

type Res struct {
	sync.Mutex
	Id     Id
	Name   string
	Flag   uint64
	Parent *Res
	*Dir
}

func (r *Res) Path() string {
	if r.Dir != nil {
		return r.Dir.Path
	}
	if r.Parent == nil {
		return r.Name
	}
	if r.Parent.Parent == nil {
		return r.Parent.Name + r.Name
	}
	return r.Parent.Path() + string(os.PathSeparator) + r.Name
}
func newChild(pa *Res, fi os.FileInfo) *Res {
	r := &Res{Name: fi.Name(), Parent: pa}
	p := r.Path()
	r.Id = NewId(p)
	if fi.IsDir() {
		r.Flag |= FlagDir
		r.Dir = &Dir{Path: p}
	}
	return r
}

type byTypeAndName []*Res

func (l byTypeAndName) Len() int {
	return len(l)
}
func less(i, j *Res) bool {
	if isdir := i.Flag&FlagDir != 0; isdir != (j.Flag&FlagDir != 0) {
		return isdir
	}
	return i.Name < j.Name
}
func (l byTypeAndName) Less(i, j int) bool {
	return less(l[i], l[j])
}
func (l byTypeAndName) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func insert(l []*Res, r *Res) []*Res {
	i := sort.Search(len(l), func(i int) bool {
		return less(r, l[i])
	})
	if i < len(l) {
		if i > 0 && l[i-1].Id == r.Id {
			l[i-1] = r
			return l
		}
		return append(l[:i], append([]*Res{r}, l[i:]...)...)
	}
	return append(l, r)
}
func remove(l []*Res, r *Res) []*Res {
	i := sort.Search(len(l), func(i int) bool {
		return less(r, l[i])
	})
	if i > 0 && l[i-1].Id == r.Id {
		return append(l[:i-1], l[i:]...)
	}
	return l
}
func find(l []*Res, name string) *Res {
	for _, r := range l {
		if r.Name == name {
			return r
		}
	}
	return nil
}

var Skip = fmt.Errorf("skip")

func walk(l []*Res, f func(*Res) error) error {
	var err error
	for _, c := range l {
		if err = f(c); err == Skip {
			continue
		}
		if err != nil {
			return err
		}
		if c.Dir != nil {
			if err = walk(c.Children, f); err != nil {
				return err
			}
		}
	}
	return nil
}
