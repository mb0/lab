// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

// Skip is returned by walk visitors to prevent visiting children of the resource in context.
var Skip = fmt.Errorf("skip")

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
	if len(r.DirPath) < 2 {
		return r.DirPath + r.Name
	}
	return r.DirPath + string(os.PathSeparator) + r.Name
}

func newChild(pa *Res, name string, isdir, stat bool) (*Res, error) {
	r := &Res{Name: name, Parent: pa, DirPath: pa.Dir.Path}
	path := r.DirPath
	if len(path) >= 2 {
		path += string(os.PathSeparator)
	}
	path += r.Name
	r.Id = NewId(path)
	if stat {
		fi, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		isdir = fi.IsDir()
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

func walk(l []*Res, f func(*Res) error) error {
	var err error
	for _, c := range l {
		if err = f(c); err == Skip {
			continue
		}
		if err != nil {
			return err
		}
		var cl []*Res
		c.Lock()
		if c.Dir != nil {
			cl = c.Children
		}
		c.Unlock()
		if len(cl) > 0 {
			if err = walk(c.Children, f); err != nil {
				return err
			}
		}
	}
	return nil
}
