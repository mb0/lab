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
	"time"

	"github.com/mb0/lab/ws"
)

func Scan(p *Pkg) error {
	src, test := getinfo(p)
	var err error
	if src != nil {
		p.Name, err = parse(p, src, "")
		src.Merge(p.Src.Info)
	}
	if test != nil {
		p.Name, err = parse(p, test, p.Name)
		test.Merge(p.Test.Info)
	}
	p.Src.Info, p.Test.Info = src, test
	return err
}
func getinfo(p *Pkg) (src, test *Info) {
	p.Res.Lock()
	defer p.Res.Unlock()
	if p.Res.Dir == nil {
		return
	}
	p.Dir = p.Res.Dir.Path
	for _, c := range p.Res.Children {
		if c.Flag&(ws.FlagDir|FlagGo) == FlagGo {
			if strings.HasSuffix(c.Name, "_test.go") {
				if test == nil {
					test = &Info{Time: time.Now().Unix()}
				}
				test.AddFile(c.Id, c.Name)
			} else {
				if src == nil {
					src = &Info{Time: time.Now().Unix()}
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
		var ok bool
		if name, ok = checkname(f.Name.Name, name); !ok {
			lasterr = fmt.Errorf("package name err: %s %s", f.Name.Name, name)
			file.Err = lasterr
			continue
		}
		for _, s := range f.Imports {
			info.AddImport(s.Path.Value[1 : len(s.Path.Value)-1])
		}
	}
	return name, lasterr
}
func checkname(name string, now string) (string, bool) {
	if strings.HasSuffix(name, "_test") {
		name = name[:len(name)-5]
	}
	if now == "" || name == now {
		return name, name != ""
	}
	return now, false
}
