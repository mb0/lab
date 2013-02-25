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
	Found PkgFlag = 1 << iota

	Scanned

	HasSource
	HasTest
	HasXTest

	Watching
	Recursing
)

type Pkg struct {
	sync.Mutex
	Id   ws.Id
	Path string
	Name string
	Flag PkgFlag
	Pkgs []*Pkg
}

type Src struct {
	sync.RWMutex
	srcids []ws.Id
	pkgs   map[ws.Id]*Pkg
	queue  *ws.Throttle
	rmchan chan *ws.Res
	w      *ws.Ws
}

func New(srcids []ws.Id) *Src {
	return &Src{
		srcids: srcids,
		pkgs:   make(map[ws.Id]*Pkg),
		queue: &ws.Throttle{
			Tickers: make(chan *time.Ticker, 1),
			Ticks:   time.Second / 2,
		},
		rmchan: make(chan *ws.Res),
	}
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
			if timeout = nil; t != nil {
				timeout = t.C
			}
		case <-timeout:
			for _, r := range s.queue.Work() {
				s.change(r)
			}
		case r := <-s.rmchan:
			s.queue.Delete(r)
			s.remove(r)
		}
	}
}

func (s *Src) WorkOn(path string) error {
	flag := Watching
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
			s.work(pkg, r)
		}
	}
	return nil
}

func (s *Src) work(p *Pkg, r *ws.Res) error {
	Scan(p, r)
	if p.Flag&Recursing != 0 {
		for _, c := range p.Pkgs {
			c.Flag |= Watching | Recursing
			cr := getchild(r, c.Id)
			if cr != nil {
				s.work(c, cr)
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
func scanpath(r *ws.Res, buf []byte) []byte {
	if r.Parent.Parent.Flag&FlagGo != 0 {
		buf = scanpath(r.Parent, buf)
		buf = append(buf, '/')
	}
	return append(buf, []byte(r.Name)...)
}
func importPath(r *ws.Res) string {
	return string(scanpath(r, make([]byte, 0, 64)))
}
func (s *Src) change(r *ws.Res) {
	s.Lock()
	defer s.Unlock()
	p := s.getorcreate(r)
	if p.Flag&Watching != 0 {
		// enqueue build and test run
		s.work(p, r)
	}
}
func (s *Src) getorcreate(r *ws.Res) *Pkg {
	p, ok := s.pkgs[r.Id]
	if !ok {
		p = &Pkg{Id: r.Id}
		s.pkgs[r.Id] = p
	}
	if p.Flag&Found == 0 {
		p.Flag |= Found
		p.Path = importPath(r)
		if r.Parent.Parent.Flag&FlagGo != 0 {
			pa := s.getorcreate(r.Parent)
			pa.Pkgs = append(pa.Pkgs, p)
		}
	}
	return p
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
}
