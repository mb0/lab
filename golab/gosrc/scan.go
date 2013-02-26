// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/mb0/lab/ws"
)

func Scan(p *Pkg, r *ws.Res) error {
	p.Flag &^= HasSource | HasTest | HasXTest
	src, test := getinfo(p, r)
	var err error
	if src != nil {
		p.Flag |= HasSource
		p.Name, err = parse(p, src, "")
		src.Merge(p.Src)
	}
	if test != nil {
		p.Flag |= HasTest
		p.Name, err = parse(p, test, p.Name)
		test.Merge(p.Test)
	}
	p.Src, p.Test = src, test
	p.Flag |= Scanned
	return err
}
func getinfo(p *Pkg, r *ws.Res) (src, test *Info) {
	r.Lock()
	defer r.Unlock()
	if r.Dir == nil {
		return
	}
	p.Dir = r.Dir.Path
	for _, c := range r.Children {
		if c.Flag&(ws.FlagDir|FlagGo) == FlagGo {
			if strings.HasSuffix(c.Name, "_test.go") {
				if test == nil {
					test = &Info{}
				}
				test.AddFile(c.Id, c.Name)
			} else {
				if src == nil {
					src = &Info{}
				}
				src.AddFile(c.Id, c.Name)
			}
		}
	}
	return
}
func parse(p *Pkg, info *Info, name string) (string, error) {
	fset := token.NewFileSet()
	var lasterr error
	for _, file := range info.Files {
		path := filepath.Join(p.Dir, file.Name)
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.ImportsOnly)
		if err != nil {
			lasterr, file.Err = err, err
			continue
		}
		var ok, xtest bool
		if name, ok, xtest = checkname(f.Name.Name, name); !ok {
			lasterr = fmt.Errorf("package name err: %s %s", f.Name.Name, name)
			file.Err = lasterr
			continue
		} else if xtest {
			p.Flag |= HasXTest
		}
		for _, s := range f.Imports {
			info.AddImport(s.Path.Value[1 : len(s.Path.Value)-1])
		}
	}
	return name, lasterr
}
func checkname(name string, now string) (string, bool, bool) {
	var xtest bool
	if strings.HasSuffix(name, "_test") {
		xtest = true
		name = name[:len(name)-5]
	}
	if now == "" || name == now {
		return name, name != "", xtest
	}
	return now, false, xtest
}
