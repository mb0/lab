// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mb0/lab/golab/gosrc"
	"github.com/mb0/lab/golab/hub"
	"github.com/mb0/lab/ws"
)

var (
	httpstart = flag.Bool("http", false, "start http server")
	httpaddr  = flag.String("addr", "localhost:8910", "http server addr")
)

func init() {
	flag.Parse()
	if !*httpstart {
		return
	}
	log.Printf("starting http://%s/\n", *httpaddr)
	serveclient()
	lab.mods = append(lab.mods, &httpmod{})
}

type httpmod struct {
	*hub.Hub
}

func (mod *httpmod) Start() {
	mod.Hub = hub.New(func(h *hub.Hub, m hub.Msg, id hub.Id) {
		switch m.Head {
		case hub.Signon.Head:
			// send reports for all working packages
			msg, err := hub.Marshal("reports", lab.src.AllReports())
			if err != nil {
				log.Println(err)
				return
			}
			h.SendMsg(msg, id)
		default:
			// echo messages
			h.SendMsg(m, id)
		}
	})
	lab.src.SignalReports(func(r *gosrc.Report) {
		m, err := hub.Marshal("report", r)
		if err != nil {
			log.Println(err)
			return
		}
		mod.SendMsg(m, hub.Group)
	})
	http.Handle("/ws", mod.Hub)
	err := http.ListenAndServe(*httpaddr, nil)
	if err != nil {
		log.Fatalf("http %s\n", err)
	}
}

func (mod *httpmod) Filter(r *ws.Res) bool {
	return false
}

func (mod *httpmod) Handle(op ws.Op, r *ws.Res) {
}

func serveclient() {
	http.HandleFunc("/", index)
	if static := findstatic(); static == "" {
		indexbytes = []byte("cannot find client files.")
	} else {
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(static))))
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method nod allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexbytes)
}

var indexbytes = []byte(`<!DOCTYPE html>
<html lang="en"><head>
	<title>golab</title>
	<meta charset="utf-8">
	<link href="/static/main.css" rel="stylesheet">
</head><body>
	<div id="app"></div>
	<script data-main="/static/main" src="/static/require.js"></script>
</body></html>
`)

func findstatic() string {
	for _, dir := range lab.dirs {
		path := filepath.Join(dir, "github.com/mb0/lab/golab/static")
		_, err := os.Stat(path)
		if err == nil {
			return path
		}
	}
	return ""
}
