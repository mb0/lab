// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"bytes"
	"fmt"
	"time"
)

func Deps(src *Src, pkg *Pkg) *Result {
	if pkg.Src.Info == nil {
		// not scanned
		return nil
	}

	retry := pkg.Flag&MissingDeps != 0
	missing := make(map[string]struct{}, 100)
	deps(src, pkg, missing)
	if len(missing) == 0 {
		// all deps found
		pkg.Flag &^= MissingDeps
		return nil
	}
	if !retry {
		// flag missing dependencies and retry later
		pkg.Flag |= MissingDeps
		return nil
	}
	var buf bytes.Buffer
	for pkgpath := range missing {
		fmt.Fprintf(&buf, "\t%s\n", pkgpath)
	}
	return &Result{
		Mode:   "deps",
		Time:   time.Now().Unix(),
		Errmsg: "missing dependencies",
		Stdout: buf.String(),
	}
}

func deps(src *Src, pkg *Pkg, missing map[string]struct{}) {
	info := pkg.Src.Info
	if info == nil {
		return
	}
	for i := range info.Imports {
		imprt := &info.Imports[i]
		p := src.lookup[imprt.Path]
		if p == nil || p.Path == "" {
			missing[imprt.Path] = struct{}{}
			continue
		}
		imprt.Id = p.Id
		p.AddUse(pkg.Id)
		if p.Flag&Watching == 0 {
			p.Flag |= Watching
		}
		if p.Src.Info == nil {
			Scan(p)
		}
		deps(src, p, missing)
	}
}
