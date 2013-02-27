// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/mb0/lab/ws"
)

var FlagGo uint64 = 1 << 16

type PkgFlag uint64

const (
	_ PkgFlag = 1 << iota

	Working
	Recursing
	Watching

	Found   // Path, Dir
	Scanned // Name, Src, Test

	HasSource
	HasTest
	HasXTest

	MissingDeps
)

type Pkg struct {
	sync.Mutex
	Id   ws.Id
	Flag PkgFlag
	Pkgs []*Pkg

	Path string
	Dir  string

	Name string
	Src  *Info
	Test *Info
}

type Src struct {
	sync.RWMutex
	srcids []ws.Id
	pkgs   map[ws.Id]*Pkg
	lookup map[string]*Pkg
	queue  *ws.Throttle
	rmchan chan *ws.Res
	w      *ws.Ws
}

func New(srcids []ws.Id) *Src {
	s := &Src{
		srcids: srcids,
		pkgs:   make(map[ws.Id]*Pkg),
		lookup: make(map[string]*Pkg),
		queue: &ws.Throttle{
			Tickers: make(chan *time.Ticker, 1),
			Ticks:   time.Second / 2,
		},
		rmchan: make(chan *ws.Res),
	}
	s.lookup["C"] = &Pkg{Flag: Found | Scanned, Path: "C", Name: "C"}
	return s
}

func (s *Src) Pkg(id ws.Id) *Pkg {
	s.Lock()
	defer s.Unlock()
	return s.pkgs[id]
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

func (s *Src) Run(w *ws.Ws) {
	s.w = w
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
	dirty := make(map[ws.Id]bool)
	s.Lock()
	for _, r := range batch {
		p := s.getorcreate(r)
		if p.Flag&Watching != 0 {
			workAll(s, p, r, dirty)
		}
	}
	s.Unlock()
	for id, dirt := range dirty {
		if dirt {
			if r := s.w.Res(id); r != nil {
				s.queue.Add(r)
			}
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
}
func (s *Src) getorcreate(r *ws.Res) *Pkg {
	p, ok := s.pkgs[r.Id]
	if !ok {
		p = &Pkg{Id: r.Id, Dir: r.Path()}
		s.pkgs[p.Id] = p
	}
	if p.Flag&Found == 0 {
		p.Flag |= Found
		p.Path = importPath(r)
		if s.lookup[p.Path] == nil {
			s.lookup[p.Path] = p
		}
		if r.Parent.Parent.Flag&FlagGo != 0 {
			pa := s.getorcreate(r.Parent)
			if pa.Flag&Recursing != 0 {
				p.Flag |= Working | Watching | Recursing
			}
			pa.Pkgs = append(pa.Pkgs, p)
		}
	}
	return p
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
	pkg := s.pkgs[id]
	if pkg == nil {
		pkg := &Pkg{Flag: flag}
		pkg.Id = id
		s.pkgs[id] = pkg
		fmt.Printf("watching %s\n", p)
		return nil
	}
	pkg.Flag |= flag
	if s.w != nil {
		if r := s.w.Res(pkg.Id); r != nil {
			workAll(s, pkg, r, make(map[ws.Id]bool))
		}
	}
	return nil
}

func work(s *Src, p *Pkg, r *ws.Res, dirty map[ws.Id]bool) {
	Scan(p, r)
	isretry := p.Flag&MissingDeps != 0
	Deps(s, p, r)
	if p.Flag&MissingDeps != 0 {
		if isretry {
			fmt.Printf("package %s missing dependencies\n", p.Path)
			return
		}
		dirty[p.Id] = true
		return
	}
	dirty[p.Id] = false
	var uses []ws.Id
	if p.Src != nil {
		rep := Install(p)
		fmt.Println(rep)
		if rep.Err != nil {
			return
		}
		uses = p.Src.Uses
	}
	if p.Test != nil {
		fmt.Println(Test(p))
	}
	for _, id := range uses {
		if _, ok := dirty[id]; !ok {
			dirty[id] = true
		}
	}
}
func workAll(s *Src, p *Pkg, r *ws.Res, dirty map[ws.Id]bool) error {
	work(s, p, r, dirty)
	if p.Flag&(Working|Recursing) != 0 {
		for _, c := range p.Pkgs {
			if c.Flag&(Working|Recursing) != 0 {
				continue
			}
			c.Flag |= Working | Watching | Recursing
			if cr := getchild(r, c.Id); cr != nil {
				workAll(s, c, cr, dirty)
			}
		}
	}
	return nil
}
func getchild(r *ws.Res, id ws.Id) *ws.Res {
	for _, c := range r.Children {
		if c.Id == id {
			return c
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
