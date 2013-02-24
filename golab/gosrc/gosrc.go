package gosrc

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/mb0/lab/ws"
)

var (
	FlagGo  uint64 = 1 << 16
	FlagSrc uint64 = 1 << 17
	FlagPkg        = FlagGo | FlagSrc
)

type Pkg struct {
	sync.Mutex
	Id               ws.Id
	ImportPath, Name string
}

type Src struct {
	sync.RWMutex
	srcids []ws.Id
	pkgs   map[ws.Id]*Pkg
	queue  *ws.Throttle
	rmchan chan *ws.Res
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
			r.Parent.Flag |= FlagSrc
			r.Flag |= FlagSrc
		}
		return false
	}
	if r.Parent.Flag&FlagGo != 0 {
		if r.Name[0] != '_' {
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
	if r.Flag&FlagSrc == 0 {
		return
	}
	if r.Flag&FlagGo != 0 {
		switch op & ws.WsMask {
		case ws.Change:
			s.queue.Add(r)
		case ws.Remove:
			s.rmchan <- r
		}
		return
	}
	if op&ws.FsMask != 0 {
		s.queue.Add(r.Parent)
	}
	return
}
func (s *Src) Run() {
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
func importPath(r *ws.Res, buf []byte) []byte {
	if r.Parent.Parent.Flag&FlagGo != 0 {
		buf = importPath(r.Parent, buf)
		buf = append(buf, '/')
	}
	return append(buf, []byte(r.Name)...)
}
func (s *Src) change(r *ws.Res) {
	s.Lock()
	defer s.Unlock()
	p, ok := s.pkgs[r.Id]
	if !ok {
		p = &Pkg{Id: r.Id, ImportPath: string(importPath(r, make([]byte, 0, 64)))}
		s.pkgs[r.Id] = p
		fmt.Printf("pkg add %s\n", p.ImportPath)
	} else {
		fmt.Printf("pkg update %s\n", p.ImportPath)
	}
	// scan ?
	// enqueue build and test run
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
	fmt.Printf("pkg remove %s\n", p.ImportPath)
}
