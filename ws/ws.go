package ws

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const (
	FlagDir uint32 = 1 << iota
	FlagLogical
	FlagMount
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
	Flag   uint32
	Parent *Res
	*Dir
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

type Watcher interface {
	Watch(r *Res) error
	Close()
}

type Config struct {
	CapHint uint
	Watcher func(*Ws) (Watcher, error)
}

type Ws struct {
	sync.RWMutex
	config  Config
	root    *Res
	all     map[Id]*Res
	watcher Watcher
}

// New creates a workspace configured with c.
func New(c Config) *Ws {
	r := &Res{Id: NewId("/")}
	w := &Ws{config: c, root: r, all: make(map[Id]*Res, c.CapHint)}
	w.all[r.Id] = r
	return w
}

// Mount adds the directory tree at path to the workspace.
func (w *Ws) Mount(path string) (*Res, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("not a directory")
	}
	r, err := w.mount(path)
	if err != nil {
		return r, err
	}
	r.Lock()
	err = read(r)
	r.Unlock()
	if err != nil {
		return r, err
	}
	w.Lock()
	defer w.Unlock()
	w.addAllChildren(r)
	return r, nil
}
func (w *Ws) mount(path string) (*Res, error) {
	id := NewId(path)
	d, f := filepath.Split(path)
	w.Lock()
	defer w.Unlock()
	r, ok := w.all[id]
	if ok {
		return r, fmt.Errorf("duplicate")
	}
	r = &Res{Id: id, Name: f, Flag: FlagDir | FlagMount, Dir: &Dir{Path: path}}
	// add virtual parent
	r.Parent = w.addParent(d[:len(d)-1])
	w.all[id] = r
	return r, nil
}

// Close closes the workspace.
// The workspace and all its resources are invalid afterwards.
func (w *Ws) Close() {
	w.Lock()
	defer w.Unlock()
	if w.watcher != nil {
		w.watcher.Close()
		w.watcher = nil
	}
	// scatter garbage
	for id, r := range w.all {
		r.Parent, r.Dir = nil, nil
		delete(w.all, id)
	}
	w.all = nil
	w.root = nil
}
func (w *Ws) addParent(path string) *Res {
	id := NewId(path)
	if r, ok := w.all[id]; ok {
		return r
	}
	parent, name := filepath.Split(path)
	var pa *Res
	if len(parent) > 0 && parent != "/" {
		pa = w.addParent(parent[:len(parent)-1])
	}
	r := &Res{Id: id, Name: name, Flag: FlagDir | FlagLogical, Parent: pa}
	w.all[id] = r
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
	r.Children = children
	for _, c := range children {
		if c.Flag&FlagDir != 0 {
			if err := read(c); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
func (w *Ws) addAllChildren(r *Res) {
	for _, c := range r.Children {
		w.all[c.Id] = c
		if c.Flag&FlagDir != 0 {
			w.addAllChildren(c)
		}
	}
	if w.watcher != nil {
		err := w.watcher.Watch(r)
		if err != nil {
			fmt.Println(err)
		}
	}
}

type byTypeAndName []*Res

func (l byTypeAndName) Len() int {
	return len(l)
}
func (l byTypeAndName) Less(i, j int) bool {
	if isdir := l[i].Flag&FlagDir != 0; isdir != (l[j].Flag&FlagDir != 0) {
		return isdir
	}
	return l[i].Name < l[j].Name
}
func (l byTypeAndName) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
