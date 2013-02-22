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
	Close() error
}
type Op uint

const (
	Add Op = 1 << iota
	Change
	Remove
	_
	Create
	Modify
	Delete
	_
	WsMask Op = 0x0F
	FsMask Op = 0xF0
)

type Config struct {
	CapHint uint
	Watcher func(Controller) (Watcher, error)
	Handler func(Op, *Res)
}

func (c Config) handle(op Op, r *Res) {
	if c.Handler != nil {
		c.Handler(op, r)
	}
}

type Controller interface {
	Control(op Op, id Id, name string) error
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
	w.addAllChildren(0, r)
	return r, nil
}
func (w *Ws) mount(path string) (*Res, error) {
	id := NewId(path)
	d, f := filepath.Split(path)
	w.Lock()
	defer w.Unlock()
	if w.watcher == nil && w.config.Watcher != nil {
		watcher, err := w.config.Watcher((*ctrl)(w))
		if err != nil {
			return nil, err
		}
		w.watcher = watcher
	}
	r, ok := w.all[id]
	if ok {
		return r, fmt.Errorf("duplicate")
	}
	r = &Res{Id: id, Name: f, Flag: FlagDir | FlagMount, Dir: &Dir{Path: path}}
	// add virtual parent
	r.Parent = w.addParent(d[:len(d)-1])
	w.all[id] = r
	w.config.handle(Add, r)
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
func (w *Ws) addAllChildren(fsop Op, r *Res) {
	for _, c := range r.Children {
		w.all[c.Id] = c
		w.config.handle(fsop|Add, c)
		if c.Flag&FlagDir != 0 {
			w.addAllChildren(fsop, c)
		}
	}
	if w.watcher != nil {
		err := w.watcher.Watch(r)
		if err != nil {
			fmt.Println(err)
		}
	}
	w.config.handle(fsop|Change, r)
}

type ctrl Ws

func (w *ctrl) Control(op Op, id Id, name string) error {
	var r, p *Res
	w.Lock()
	defer w.Unlock()
	r = w.all[id]
	if name != "" {
		p, r = r, nil
		if p != nil && p.Dir != nil {
			p.Lock()
			r = find(p.Children, name)
			p.Unlock()
		}
	}
	switch {
	case op&Delete != 0:
		if r == nil {
			break
		}
		return w.remove(op, r)
	case r != nil:
		// res found, modify
		return w.change(op, r)
	case p != nil:
		// parent found create child
		return w.add(op, p, name)
	}
	// not found, ignore
	return nil
}
func (w *ctrl) change(fsop Op, r *Res) error {
	w.config.handle(fsop|Change, r)
	return nil
}
func (w *ctrl) remove(fsop Op, r *Res) error {
	if p := r.Parent; p != nil {
		p.Lock()
		defer p.Unlock()
		if p.Dir != nil {
			p.Children = remove(p.Children, r)
		}
	}
	rm := []*Res{r}
	if r.Dir != nil {
		walk(r.Children, func(c *Res) error {
			rm = append(rm, c)
			return nil
		})
	}
	for i := len(rm) - 1; i >= 0; i-- {
		c := rm[i]
		w.config.handle(fsop|Remove, c)
		delete(w.all, c.Id)
	}
	return nil
}
func (w *ctrl) add(fsop Op, p *Res, name string) error {
	p.Lock()
	defer p.Unlock()
	// new lock try again
	r := find(p.Children, name)
	// ignore duplicate
	if r != nil {
		return nil
	}
	r = &Res{Name: name, Parent: p}
	path := r.Path()
	r.Id = NewId(path)
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	p.Children = insert(p.Children, r)
	w.all[r.Id] = r
	if !fi.IsDir() {
		// send event
		return nil
	}
	r.Flag |= FlagDir
	r.Dir = &Dir{Path: path}
	if err = read(r); err != nil {
		return err
	}
	w.config.handle(fsop|Add, r)
	(*Ws)(w).addAllChildren(fsop, r)
	return nil
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
func find(l []*Res, name string) *Res {
	for _, r := range l {
		if r.Name == name {
			return r
		}
	}
	return nil
}
