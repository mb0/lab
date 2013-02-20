package ws

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"
)

const (
	FlagDir uint32 = 1 << iota
	FlagLogical
	FlagMount
)

type Res struct {
	Name   string
	Flag   uint32
	Parent *Res
	*Dir
}

type Dir struct {
	Path     string
	Children []*Res
}

func (r *Res) Path() string {
	if r == nil {
		return ""
	}
	if r.Dir != nil {
		return r.Dir.Path
	}
	return r.Parent.Path() + string(os.PathSeparator) + r.Name
}

var all = map[string]*Res{"/": &Res{Name: ""}}

func addParent(path string) *Res {
	if r, ok := all[path]; ok {
		return r
	}
	parent, name := filepath.Split(path)
	var pa *Res
	if len(parent) > 0 && parent != "/" {
		pa = addParent(parent[:len(parent)-1])
	}
	r := &Res{name, FlagDir | FlagLogical, pa, nil}
	all[path] = r
	return r
}
func newChild(pa *Res, fi os.FileInfo) *Res {
	r := &Res{fi.Name(), 0, pa, nil}
	if fi.IsDir() {
		r.Flag |= FlagDir
		r.Dir = &Dir{Path: r.Path()}
	}
	return r
}
func read(r *Res) error {
	f, err := os.Open(r.Dir.Path)
	if err != nil {
		return err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return err
	}
	children := make([]*Res, 0, len(list))
	for _, fi := range list {
		children = append(children, newChild(r, fi))
	}
	sort.Sort(byTypeAndName(children))
	r.Dir.Children = children
	for _, c := range children {
		all[c.Path()] = c
		if c.Dir != nil {
			if err := read(c); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

type byTypeAndName []*Res

func (l byTypeAndName) Len() int {
	return len(l)
}

func (l byTypeAndName) Less(i, j int) bool {
	if isdir := l[i].Dir != nil; isdir != (l[j].Dir != nil) {
		return isdir
	}
	return l[i].Name < l[j].Name
}

func (l byTypeAndName) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func Mount(path string) (*Res, error) {
	r, ok := all[path]
	if ok {
		return r, nil
	}
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("not a directory")
	}
	d, f := filepath.Split(path)
	// add virtual parent
	pa := addParent(d[:len(d)-1])
	r = &Res{f, FlagDir | FlagMount, pa, &Dir{Path: path}}
	all[path] = r
	return r, read(r)
}

func TestWalkSrc(t *testing.T) {
	dirs := build.Default.SrcDirs()
	t.Log(dirs)
	start := time.Now()
	for _, path := range dirs {
		_, err := Mount(path)
		if err != nil {
			t.Error(err)
		}
	}
	took := time.Since(start)
	runtime.GC()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	f := "count: %d, took: %s, alloc: %d/%d kb, heap: %d/%d kb, objs: %d, gcs: %d"
	kb := func(n uint64) uint64 { return n / (1 << 10) }
	t.Logf(f, len(all), took, kb(mem.Alloc), kb(mem.TotalAlloc), kb(mem.HeapAlloc), kb(mem.HeapSys), mem.HeapObjects, mem.NumGC)
	for p, r := range all {
		if p != r.Path() {
			t.Error(p, "!=", r.Path())
		}
	}
}
