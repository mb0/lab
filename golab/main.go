package main

import (
	"flag"
	"fmt"
	"github.com/mb0/lab/ws"
	"os"
	"os/signal"
	"path/filepath"
)

var paths = flag.String("paths", "", "paths to watch. defaults to cwd")

func filter(r *ws.Res) bool {
	if len(r.Name) > 0 && r.Name[0] == '.' {
		return true
	}
	return false
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
	dirs := filepath.SplitList(*paths)
	fmt.Printf("starting lab for %v\n", dirs)
	w := ws.New(ws.Config{
		CapHint: 1000,
		Watcher: ws.NewInotify,
		Filter:  filter,
		Handler: func(op ws.Op, r *ws.Res) {
			if r.Flag&ws.FlagIgnore == 0 {
				fmt.Printf("op %x for %s\n", op, r.Path())
			}
		},
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
