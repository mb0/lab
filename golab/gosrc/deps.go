// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"fmt"
	"github.com/mb0/lab/ws"
)

func Deps(src *Src, pkg *Pkg, r *ws.Res) {
	if pkg.Flag&Scanned == 0 {
		// not scanned
		return
	}

	if pkg.Src == nil {
		return
	}
	missing := make(map[string]struct{}, 100)
	deps(src, pkg, missing)
	if len(missing) > 0 {
		// if this is not already a retry
		if pkg.Flag&MissingDeps == 0 {
			// enqueue for later processing
			pkg.Flag |= MissingDeps
			src.queue.Add(r)
		} else {
			fmt.Printf("missing %v\n", missing)
		}
	} else {
		// all deps found
		pkg.Flag &^= MissingDeps
	}
}

func deps(src *Src, pkg *Pkg, missing map[string]struct{}) {
	if pkg.Src == nil {
		return
	}
	for i := range pkg.Src.Imports {
		imprt := &pkg.Src.Imports[i]
		p := src.lookup[imprt.Path]
		if p == nil || p.Flag&Found == 0 {
			missing[imprt.Path] = struct{}{}
			continue
		}
		imprt.Id = p.Id
		if p.Src == nil {
			p.Src = &Info{}
		}
		p.Src.AddUse(pkg.Id)
		if p.Flag&Watching == 0 {
			p.Flag |= Watching
		}
		if p.Flag&Scanned == 0 {
			if src.w != nil {
				if r := src.w.Res(p.Id); r != nil {
					Scan(p, r)
				}
			}
		}
		deps(src, p, missing)
	}
}
