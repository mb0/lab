package ws

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
)

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
	// CapHint hints the expected peak resource capacity.
	CapHint uint
	// Watcher returns a new watcher given workspace control.
	// Mounting a path results in a snapshot if no Watcher is configured.
	Watcher func(Controller) (Watcher, error)
	// Handler handles resource operation events if set.
	Handler func(Op, *Res)
	// Filter filters resources if set.
	// If filter returns true the resource is flagged with FlagIgnore but remains in the workspace.
	// Ignored directories are not read.
	Filter func(*Res) bool
}

func (c *Config) filldefaullts() {
	if c.Filter == nil {
		c.Filter = nilfilter
	}
	if c.Handler == nil {
		c.Handler = nilhandler
	}
}
func nilhandler(Op, *Res) {}
func nilfilter(*Res) bool { return false }

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
	c.filldefaullts()
	var name string
	if runtime.GOOS != "windows" {
		name = "/"
	}
	r := &Res{Id: NewId(name), Name: name}
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
	if w.config.Filter(r) {
		r.Flag |= FlagIgnore
		return r, nil
	}
	r.Lock()
	err = read(r, w.config.Filter)
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
	path = filepath.Clean(path)
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
	r.Parent = w.logicalParent(d)
	w.all[id] = r
	w.config.Handler(Add, r)
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
		r.Dir = nil
		delete(w.all, id)
	}
	w.all = nil
	w.root = nil
}
func (w *Ws) logicalParent(path string) *Res {
	parts := split(path)
	r := w.root
	for i := len(parts) - 1; i >= 0; i-- {
		if r.Dir == nil {
			r.Dir = &Dir{Path: r.Path()}
		} else if c := find(r.Children, parts[i]); c != nil {
			r = c
			continue
		}
		c := &Res{Name: parts[i], Parent: r, Flag: FlagDir | FlagLogical}
		p := c.Path()
		c.Dir = &Dir{Path: p}
		c.Id = NewId(p)
		r.Children = insert(r.Children, c)
		w.all[c.Id], r = c, c
	}
	return r
}
func split(path string) []string {
	parts := make([]string, 0, 8)
	dir, file := path, ""
	for dir != "" {
		if i := len(dir) - 1; dir[i] == os.PathSeparator {
			dir = dir[:i]
		}
		dir, file = filepath.Split(dir)
		if file != "" {
			parts = append(parts, file)
			continue
		}
		break
	}
	if dir != "" {
		return append(parts, dir)
	}
	return parts
}
func read(r *Res, filter func(*Res) bool) error {
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
		if filter(c) {
			c.Flag |= FlagIgnore
			continue
		}
		if c.Flag&FlagDir != 0 {
			if err := read(c, filter); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
func (w *Ws) addAllChildren(fsop Op, r *Res) {
	for _, c := range r.Children {
		w.all[c.Id] = c
		w.config.Handler(fsop|Add, c)
		if c.Flag&(FlagDir|FlagIgnore) == FlagDir {
			w.addAllChildren(fsop, c)
		}
	}
	if w.watcher != nil {
		err := w.watcher.Watch(r)
		if err != nil {
			fmt.Println(err)
		}
	}
	w.config.Handler(fsop|Change, r)
}
