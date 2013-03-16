// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
	"os/signal"

	"github.com/mb0/lab"
	"github.com/mb0/lab/golab/gosrc"
	"github.com/mb0/lab/ws"
)

type golab struct {
	roots    []string
	ws       *ws.Ws
	filters  []ws.Filter
	handlers []ws.Handler
}

func main() {
	flag.Parse()
	roots := build.Default.SrcDirs()
	lab.Register("roots", roots)
	lab.Register("gosrc", gosrc.New())
	golab := &golab{roots: roots}
	golab.ws = ws.New(ws.Config{
		CapHint: 8000,
		Watcher: ws.NewInotify,
		Filter:  golab,
		Handler: golab,
	})
	defer golab.ws.Close()
	lab.Register("ws", golab.ws)
	lab.Register("golab", golab)
	lab.Start()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}

func (l *golab) Init() {
	for _, mod := range lab.All() {
		if _, ok := mod.(*golab); ok {
			continue
		}
		if f, ok := mod.(ws.Filter); ok {
			l.filters = append(l.filters, f)
		}
		if h, ok := mod.(ws.Handler); ok {
			l.handlers = append(l.handlers, h)
		}
	}
	fmt.Println("starting lab for:", l.roots)
	for i, err := range ws.MountAll(l.ws, l.roots) {
		if err != nil {
			fmt.Printf("error mounting %s: %s\n", l.roots[i], err)
		}
	}
}

func (l *golab) Filter(r *ws.Res) bool {
	if len(r.Name) == 0 {
		return false
	}
	if r.Name[0] == '.' || r.Name[len(r.Name)-1] == '~' {
		return true
	}
	if r.Flag&ws.FlagDir == 0 && len(r.Name) > 4 {
		switch r.Name[len(r.Name)-4:] {
		case ".swp", ".swo":
			return true
		}
	}
	for _, f := range l.filters {
		if f.Filter(r) {
			return true
		}
	}
	return false
}

func (l *golab) Handle(op ws.Op, r *ws.Res) {
	if r.Flag&ws.FlagIgnore != 0 {
		return
	}
	for _, h := range l.handlers {
		h.Handle(op, r)
	}
}
