// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"github.com/mb0/lab/ws"
)

type Import struct {
	Path string
	Id   ws.Id
}
type File struct {
	Id   ws.Id
	Name string
	Err  error
}

type Info struct {
	Time    int64
	Files   []File
	Imports []Import
}

func (nfo *Info) Copy() *Info {
	if nfo == nil {
		return nil
	}
	i := *nfo
	i.Imports = make([]Import, len(nfo.Imports))
	copy(i.Imports, nfo.Imports)
	return &i
}

func (nfo *Info) Import(path string) *Import {
	for i := range nfo.Imports {
		imprt := &nfo.Imports[i]
		if path == imprt.Path {
			return imprt
		}
	}
	return nil
}
func (nfo *Info) AddImport(path string) {
	if nfo.Import(path) == nil {
		nfo.Imports = append(nfo.Imports, Import{Path: path})
	}
}
func (nfo *Info) File(id ws.Id) *File {
	for i := range nfo.Files {
		file := &nfo.Files[i]
		if id == file.Id {
			return file
		}
	}
	return nil
}
func (nfo *Info) AddFile(id ws.Id, name string) {
	if nfo.File(id) == nil {
		nfo.Files = append(nfo.Files, File{Id: id, Name: name})
	}
}
func (nfo *Info) Merge(old *Info) {
	if nfo == nil || old == nil {
		return
	}
	for i := range nfo.Imports {
		imprt := &nfo.Imports[i]
		if imprt.Id == 0 {
			with := old.Import(imprt.Path)
			if with != nil {
				imprt.Id = with.Id
			}
		}
	}
	return
}
