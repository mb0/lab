// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
)

const (
	FlagDir uint64 = 1 << iota
	FlagLogical
	FlagMount
	FlagIgnore
)

// Id identifies a workspace resource uniquely.
// Having a fnv32 hash collision is considered a user error.
type Id uint32

// Creates a workspace id for path.
// Path must be absolute and clean (sans trailing slash).
func NewId(path string) Id {
	h := fnv.New32()
	h.Write([]byte(path))
	return Id(h.Sum32())
}
func (id Id) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`"%X"`, id)
	return []byte(str), nil
}
func (id *Id) UnmarshalJSON(data []byte) error {
	_, err := fmt.Sscanf(string(data), `"%X"`, id)
	return err
}

// Dir holds the full path and child resources for a directory resource.
type Dir struct {
	Path     string
	Children []*Res
}

// Res describes a workspace resource.
type Res struct {
	sync.Mutex
	Id      Id
	Name    string
	Flag    uint64
	Parent  *Res
	DirPath string
	*Dir
}

// Path returns the full resource path.
func (r *Res) Path() string {
	r.Lock()
	defer r.Unlock()
	if r.Dir != nil {
		return r.Dir.Path
	}
	return r.DirPath + r.Name
}

func (w *Ws) newChild(pa *Res, name string, isdir, stat bool) (*Res, error) {
	path := pa.Dir.Path
	if len(path) >= 2 {
		path += w.fs.Seperator
	}
	r := &Res{Name: name, Parent: pa, DirPath: path}
	path += r.Name
	r.Id = NewId(path)
	if stat {
		var err error
		isdir, err = w.fs.IsDir(path)
		if err != nil {
			return nil, err
		}
	}
	if isdir {
		r.Flag |= FlagDir
		r.Dir = &Dir{Path: path}
	}
	return r, nil
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
