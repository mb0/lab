// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

func Deps(src *Src, pkg *Pkg) {
	if pkg.Src.Info == nil {
		// not scanned
		return
	}

	missing := make(map[string]struct{}, 100)
	deps(src, pkg, missing)
	if len(missing) > 0 {
		pkg.Flag |= MissingDeps
	} else {
		// all deps found
		pkg.Flag &^= MissingDeps
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
