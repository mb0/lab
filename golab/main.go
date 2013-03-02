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
	"path/filepath"

	"github.com/mb0/lab/golab/gosrc"
	"github.com/mb0/lab/ws"
)

var workpaths = flag.String("work", "./...", "path list of active packages. defaults to cwd")

type Module interface {
	Start()
	Filter(*ws.Res) bool
	Handle(ws.Op, *ws.Res)
}

var lab = &struct {
	dirs []string
	src  *gosrc.Src
	ws   *ws.Ws
	mods []Module
}{
	dirs: build.Default.SrcDirs(),
}

func filter(r *ws.Res) bool {
	l := len(r.Name)
	if l == 0 {
		return false
	}
	if r.Name[0] == '.' || r.Name[l-1] == '~' {
		return true
	}
	if r.Flag&ws.FlagDir == 0 && l > 4 {
		switch r.Name[l-4:] {
		case ".swp", ".swo":
			return true
		}
	}
	for _, mod := range lab.mods {
		if mod.Filter(r) {
			return true
		}
	}
	return false
}

func handler(op ws.Op, r *ws.Res) {
	if r.Flag&ws.FlagIgnore != 0 {
		return
	}
	for _, mod := range lab.mods {
		mod.Handle(op, r)
	}
}

func main() {
	flag.Parse()
	ids := make([]ws.Id, 0, len(lab.dirs))
	for _, d := range lab.dirs {
		ids = append(ids, ws.NewId(d))
	}
	lab.src = gosrc.New(ids)
	lab.mods = append(lab.mods, lab.src)
	lab.src.SignalReports(func(r *gosrc.Report) {
		fmt.Println(r)
	})
	lab.ws = ws.New(ws.Config{
		CapHint: 8000,
		Watcher: ws.NewInotify,
		Filter:  filter,
		Handler: handler,
	})
	defer lab.ws.Close()
	for _, p := range filepath.SplitList(*workpaths) {
		err := lab.src.WorkOn(p)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	fmt.Println("starting lab for:", lab.dirs)
	for i, err := range ws.MountAll(lab.ws, lab.dirs) {
		if err != nil {
			fmt.Printf("error mounting %s: %s\n", lab.dirs[i], err)
		}
	}
	for _, mod := range lab.mods {
		go mod.Start()
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
