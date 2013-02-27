// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/mb0/lab/golab/gosrc"
	"github.com/mb0/lab/golab/hub"
)

var (
	httpstart = flag.Bool("http", false, "start http server")
	httpaddr  = flag.String("addr", "localhost:8910", "http server addr")
)

func router(h *hub.Hub, m hub.Msg, id hub.Id) {
	// echo messages
	h.SendMsg(m, id)
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

func init() {
	flag.Parse()
	if !*httpstart {
		return
	}
	modules = append(modules, mainhttp)
}

func mainhttp() {
	h := hub.New(router)
	lab.src.SignalReports(func(r *gosrc.Report) {
		m, err := hub.Marshal("report", r)
		if err != nil {
			log.Println(err)
			return
		}
		h.SendMsg(m, hub.Group)
	})
	http.Handle("/ws", h)
	http.HandleFunc("/", index)
	log.Printf("starting http://%s/\n", *httpaddr)
	err := http.ListenAndServe(*httpaddr, nil)
	if err != nil {
		log.Fatalf("http %s\n", err)
	}
}

var indexbytes = []byte(`<!DOCTYPE html>
<html lang="en"><head>
	<meta charset="utf-8">
	<title>golab</title>
</head><body><script>(function() {
function newchild(pa, tag, inner) {
	var ele = document.createElement(tag);
	pa.appendChild(ele);
	if (inner) ele.innerHTML = inner;
	return ele;
}
var cont = newchild(document.body, "div");
if (window["WebSocket"]) {
	var conn = new WebSocket("ws://"+ location.host+"/ws");
	conn.onclose = function(e) {
		newchild(cont, "p", "WebSocket closed.");
	};
	conn.onmessage = function(e) {
		newchild(cont, "p", e.data);
	};
	conn.onopen = function(e) {
		newchild(cont, "p", "WebSocket started.");
		conn.send('{"Head":"hi"}\n');
	};
} else {
	newchild(cont, "p", "WebSockets are not supported by your browser.");
}
})()</script></body></html>
`)
