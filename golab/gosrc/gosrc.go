// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"fmt"
	"go/build"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/mb0/lab"
	"github.com/mb0/lab/ws"
)

var workpaths = lab.Conf.String("work", "./...", "path list of active packages. defaults to cwd")

var FlagGo uint64 = 1 << 16

type Src struct {
	sync.RWMutex
	srcids []ws.Id
	pkgs   map[ws.Id]*Pkg
	lookup map[string]*Pkg
	queue  *ws.Throttle
	rmchan chan *ws.Res

	reportsignal []func(*Report)
}

func New() *Src {
	s := &Src{
		pkgs:   make(map[ws.Id]*Pkg),
		lookup: make(map[string]*Pkg),
		queue:  ws.NewThrottle(time.Second),
		rmchan: make(chan *ws.Res),
	}
	p := Pkg{Id: ws.NewId("C"), Path: "C"}
	p.Name = "C"
	p.Src.Info = &Info{}
	s.lookup["C"] = &p
	return s
}

func (s *Src) Init() {
	var ids []ws.Id
	for _, d := range build.Default.SrcDirs() {
		ids = append(ids, ws.NewId(d))
	}
	s.srcids = ids
	for _, p := range filepath.SplitList(*workpaths) {
		err := s.WorkOn(p)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	s.SignalReports(func(r *Report) {
		fmt.Println(r)
	})
}

func (s *Src) Pkg(id ws.Id) *Pkg {
	s.Lock()
	defer s.Unlock()
	return s.pkgs[id]
}

func (s *Src) Find(path string) *Pkg {
	s.Lock()
	defer s.Unlock()
	return s.lookup[path]
}

func (s *Src) SignalReports(f func(*Report)) {
	s.Lock()
	defer s.Unlock()
	s.reportsignal = append(s.reportsignal, f)
}

func (s *Src) AllReports() []*Report {
	s.Lock()
	defer s.Unlock()
	var reps []*Report
	for _, pkg := range s.lookup {
		if pkg.Flag&Working != 0 && (pkg.Src.Info != nil || pkg.Test.Info != nil) {
			reps = append(reps, NewReport(pkg))
		}
	}
	sort.Sort(byDir(reps))
	return reps
}

func (s *Src) Filter(r *ws.Res) bool {
	if r.Flag&ws.FlagDir == 0 {
		if filepath.Ext(r.Name) == ".go" {
			r.Flag |= FlagGo
		}
		return false
	}
	if r.Parent.Flag&FlagGo != 0 {
		if r.Name != "testdata" && r.Name[0] != '_' {
			r.Flag |= FlagGo
		}
		return false
	}
	if r.Name == "pkg" || r.Name == "src" {
		for _, id := range s.srcids {
			if r.Id == id {
				r.Flag |= FlagGo
				break
			}
		}
	}
	return false
}

func (s *Src) Handle(op ws.Op, r *ws.Res) {
	if r.Flag&FlagGo == 0 {
		return
	}
	if r.Flag&ws.FlagDir != 0 {
		switch op & ws.WsMask {
		case ws.Change:
			s.queue.Add(r)
		case ws.Remove:
			s.rmchan <- r
		}
		return
	}
	if op&ws.FsMask != 0 && r.Parent.Flag&FlagGo != 0 {
		s.queue.Add(r.Parent)
	}
	return
}

func (s *Src) Run() {
	var timeout <-chan time.Time
	for {
		select {
		case t := <-s.queue.Tickers:
			timeout = t.C
		case <-timeout:
			s.change(s.queue.Work())
		case r := <-s.rmchan:
			s.queue.Delete(r)
			s.remove(r)
		}
	}
}
func (s *Src) change(batch []*ws.Res) {
	dirty := make(map[ws.Id]*Pkg)
	s.Lock()
	for _, r := range batch {
		p := s.getorcreateres(r)
		if p.Flag&Watching != 0 {
			workAll(s, p, dirty)
		}
	}
	for _, dirt := range dirty {
		if dirt != nil {
			workAll(s, dirt, dirty)
		}
	}
	s.Unlock()
	for _, dirt := range dirty {
		if dirt != nil {
			s.queue.Add(dirt.Res)
		}
	}
}
func (s *Src) remove(r *ws.Res) {
	s.Lock()
	defer s.Unlock()
	p, ok := s.pkgs[r.Id]
	if !ok {
		return
	}
	// clean up p
	delete(s.pkgs, p.Id)
	delete(s.lookup, p.Path)
	// TODO clean up uses in dependencies
}
func (s *Src) getorcreate(id ws.Id, dir string) *Pkg {
	pkg, ok := s.pkgs[id]
	if !ok {
		pkg = &Pkg{Id: id, Dir: dir}
		s.pkgs[id] = pkg
	}
	return pkg
}
func (s *Src) getorcreateres(r *ws.Res) *Pkg {
	pkg := s.getorcreate(r.Id, r.Path())
	if pkg.Res == nil {
		pkg.Res = r
		pkg.Path = importPath(r)
		if s.lookup[pkg.Path] == nil {
			s.lookup[pkg.Path] = pkg
		}
		if r.Parent.Parent.Flag&FlagGo != 0 {
			pa := s.getorcreateres(r.Parent)
			if pa.Flag&Recursing != 0 {
				pkg.Flag |= Working | Watching | Recursing
			}
			pa.Pkgs = append(pa.Pkgs, pkg)
		}
	}
	return pkg
}

func (s *Src) WorkOn(path string) error {
	flag := Working | Watching
	if d, f := filepath.Split(path); f == "..." {
		path = d
		flag |= Recursing
	}
	p, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	id := ws.NewId(p)
	s.Lock()
	defer s.Unlock()
	pkg := s.getorcreate(id, p)
	pkg.Flag |= flag
	if pkg.Res != nil {
		workAll(s, pkg, make(map[ws.Id]*Pkg))
	}
	return nil
}

func work(s *Src, p *Pkg, dirty map[ws.Id]*Pkg) {
	Scan(p)
	isretry := p.Flag&MissingDeps != 0
	Deps(s, p)
	if p.Flag&MissingDeps != 0 {
		if isretry {
			fmt.Printf("package %s missing dependencies\n", p.Path)
			return
		}
		dirty[p.Id] = p
		return
	}
	dirty[p.Id] = nil
	if p.Src.Info != nil {
		p.Src.Result = Install(p)
	}
	if p.Test.Info != nil {
		p.Test.Result = Test(p)
	}
	if p.Src.Result != nil || p.Test.Result != nil {
		rep := NewReport(p)
		for _, f := range s.reportsignal {
			f(rep)
		}
	}
	for _, id := range p.Uses {
		if _, ok := dirty[id]; !ok {
			dirty[id] = s.pkgs[id]
		}
	}
}
func workAll(s *Src, p *Pkg, dirty map[ws.Id]*Pkg) error {
	work(s, p, dirty)
	if p.Flag&(Working|Recursing) != 0 {
		for _, c := range p.Pkgs {
			if c.Res == nil || c.Flag&(Working|Recursing) != 0 {
				continue
			}
			c.Flag |= Working | Watching | Recursing
			workAll(s, c, dirty)
		}
	}
	return nil
}

func importPath(r *ws.Res) string {
	return string(scanpath(r, make([]byte, 0, 64)))
}
func scanpath(r *ws.Res, buf []byte) []byte {
	if r.Parent.Parent.Flag&FlagGo != 0 {
		buf = scanpath(r.Parent, buf)
		buf = append(buf, '/')
	}
	return append(buf, []byte(r.Name)...)
}
