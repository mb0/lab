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

var modules []func()
var lab = &struct {
	dirs []string
	src  *gosrc.Src
	ws   *ws.Ws
}{}

func filter(r *ws.Res) bool {
	if len(r.Name) > 0 && r.Name[0] == '.' {
		return true
	}
	return lab.src.Filter(r)
}
func handler(op ws.Op, r *ws.Res) {
	if r.Flag&ws.FlagIgnore != 0 {
		return
	}
	lab.src.Handle(op, r)
}
func main() {
	flag.Parse()
	lab.dirs = build.Default.SrcDirs()
	ids := make([]ws.Id, 0, len(lab.dirs))
	for _, d := range lab.dirs {
		ids = append(ids, ws.NewId(d))
	}
	lab.src = gosrc.New(ids)
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
	for _, module := range modules {
		go module()
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
