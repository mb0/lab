// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"sync"

	"github.com/mb0/lab/ws"
)

type PkgFlag uint64

const (
	_ PkgFlag = 1 << iota

	Working
	Recursing
	Watching
	MissingDeps
)

type Pkg struct {
	sync.Mutex
	Id   ws.Id
	Flag PkgFlag
	Pkgs []*Pkg

	// Found
	Res *ws.Res
	Dir string

	// Valid
	Path string

	// Scanned
	Detail
}

type Code struct {
	Info   *Info   `json:",omitempty"`
	Result *Result `json:",omitempty"`
}

type Detail struct {
	Name string
	Src  Code `json:",omitempty"`
	Test Code `json:",omitempty"`
	Uses []ws.Id
}

func (d *Detail) AddUse(id ws.Id) {
	for _, i := range d.Uses {
		if i == id {
			return
		}
	}
	d.Uses = append(d.Uses, id)
}
