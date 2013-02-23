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

var paths = flag.String("paths", "", "paths to watch. defaults to cwd")

var src *gosrc.Src
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
	var err error
	if *paths == "" {
		*paths, err = os.Getwd()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	// TODO create active working set
	_ = ids(filepath.SplitList(*paths))
	dirs := build.Default.SrcDirs()
	src = gosrc.New(ids(dirs))
	go src.Run()
	fmt.Printf("starting lab for %v\n", dirs)
	w := ws.New(ws.Config{
		CapHint: 8000,
		Watcher: ws.NewInotify,
		Filter:  filter,
		Handler: handler,
	})
	defer w.Close()
	for _, dir := range dirs {
		_, err = w.Mount(dir)
		if err != nil {
			fmt.Println(dir, err)
			return
		}
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
