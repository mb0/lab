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
var dirs = build.Default.SrcDirs()
var src = gosrc.New(ids(dirs))

func filter(r *ws.Res) bool {
	if len(r.Name) > 0 && r.Name[0] == '.' {
		return true
	}
	return src.Filter(r)
}
func handler(op ws.Op, r *ws.Res) {
	if r.Flag&ws.FlagIgnore != 0 {
		return
	}
	src.Handle(op, r)
}
func ids(paths []string) []ws.Id {
	ids := make([]ws.Id, 0, len(paths))
	for _, p := range paths {
		ids = append(ids, ws.NewId(p))
	}
	return ids
}
func main() {
	flag.Parse()
	initwork(*workpaths)
	fmt.Printf("starting lab for %v\n", dirs)
	w := ws.New(ws.Config{
		CapHint: 8000,
		Watcher: ws.NewInotify,
		Filter:  filter,
		Handler: handler,
	})
	defer w.Close()
	for i, err := range ws.MountAll(w, dirs) {
		if err != nil {
			fmt.Printf("error mounting %s: %s\n", dirs[i], err)
		}
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
func initwork(paths string) {
	list := filepath.SplitList(paths)
	for _, p := range list {
		err := src.WorkOn(p)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
