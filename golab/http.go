// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"regexp"

	_ "github.com/mb0/ace"
	"github.com/mb0/lab"
	"github.com/mb0/lab/golab/gosrc"
	"github.com/mb0/lab/hub"
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
	lab.Register("http", &httpmod{})
}

type httpmod struct {
	*hub.Hub
}

func (mod *httpmod) Run() {
	src := lab.Mod("gosrc").(*gosrc.Src)
	mod.Hub = hub.New(func(h *hub.Hub, m hub.Msg, id hub.Id) {
		switch m.Head {
		case hub.Signon.Head:
			// send reports for all working packages
			msg, err := hub.Marshal("reports", src.AllReports())
			if err != nil {
				log.Println(err)
				return
			}
			h.SendMsg(msg, id)
		case "stat":
			var path string
			err := m.Unmarshal(&path)
			if err != nil {
				log.Println(err)
				return
			}
			msg, err := apistat(path)
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
	src.SignalReports(func(r *gosrc.Report) {
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

func apistat(path string) (hub.Msg, error) {
	res := apires{ws.NewId(path), path, false}
	labws := lab.Mod("ws").(*ws.Ws)
	if r := labws.Res(res.Id); r != nil {
		r.Lock()
		defer r.Unlock()
		res.Name, res.IsDir = r.Name, r.Flag&ws.FlagDir != 0
		if r.Dir != nil {
			cs := make([]apires, 0, len(r.Children))
			for _, c := range r.Children {
				if c.Flag&ws.FlagIgnore == 0 {
					cs = append(cs, apires{c.Id, c.Name, c.Flag&ws.FlagDir != 0})
				}
			}
			return hub.Marshal("stat", struct {
				apires
				Path     string
				Children []apires
			}{res, path, cs})
		}
		return hub.Marshal("stat", struct {
			apires
			Path string
		}{res, path})
	}
	return hub.Marshal("stat.err", struct {
		apires
		Path  string
		Error string
	}{res, path, "not found"})
}

type apires struct {
	Id    ws.Id `json:",string"`
	Name  string
	IsDir bool
}

func serveclient() {
	http.HandleFunc("/", index)
	static, staticace := findstatic(), findace()
	if static == "" || staticace == "" {
		indexbytes = []byte("cannot find client files.")
	} else {
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(static))))
		http.Handle("/static/ace/", http.StripPrefix("/static/ace/", http.FileServer(http.Dir(staticace))))
		http.HandleFunc("/manifest", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/cache-manifest; charset=utf-8")
			http.ServeFile(w, r, filepath.Join(static, "main.manifest"))
		})
	}
	http.HandleFunc("/raw/", raw)
	http.HandleFunc("/doc/", godoc)
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
<html lang="en" manifest="/manifest"><head>
	<title>golab</title>
	<meta charset="utf-8">
	<link href="/static/main.css" rel="stylesheet">
</head><body>
	<div id="app"></div>
	<script data-main="/static/main" src="http://cdnjs.cloudflare.com/ajax/libs/require.js/2.1.4/require.min.js"></script>
</body></html>
`)

func findstatic() string {
	roots := lab.Mod("roots").([]string)
	for _, dir := range roots {
		path := filepath.Join(dir, "github.com/mb0/lab/golab/static")
		_, err := os.Stat(path)
		if err == nil {
			return path
		}
	}
	return ""
}

func findace() string {
	roots := lab.Mod("roots").([]string)
	for _, dir := range roots {
		path := filepath.Join(dir, "github.com/mb0/ace/lib/ace")
		_, err := os.Stat(path)
		if err == nil {
			return path
		}
	}
	return ""
}

func raw(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[:5] != "/raw/" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case "GET", "POST":
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path[4:]
	labws := lab.Mod("ws").(*ws.Ws)
	if res := labws.Res(ws.NewId(path)); res == nil || res.Flag&(ws.FlagDir|ws.FlagIgnore) != 0 {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case "GET":
		f, err := os.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err = io.Copy(w, f)
		if err != nil {
			log.Println(err)
		}
	case "POST":
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		_, err = io.Copy(f, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func godoc(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[:5] != "/doc/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path[5:]
	src := lab.Mod("gosrc").(*gosrc.Src)
	pkg := src.Find(path)
	if pkg == nil {
		http.NotFound(w, r)
		return
	}
	pkg.Lock()
	dir := pkg.Dir
	pkg.Unlock()
	raw, err := gosrc.LoadHtmlDoc(path, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// fix source links
	regex, err := regexp.Compile(fmt.Sprintf(`<a href="(/src/pkg/%s)(.*?\.go)(\?s=\d+:\d+(#L\d+))?"`, path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	raw = regex.ReplaceAll(raw, []byte(fmt.Sprintf(`<a href="#file%s$2$4"`, dir)))
	w.Write(raw)
}
