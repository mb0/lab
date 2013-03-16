// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package htmod

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/mb0/lab/golab/gosrc"
	"github.com/mb0/lab/ws"
)

func (mod *htmod) serveContent() {
	http.Handle("/raw/", (*srvraw)(mod))
	http.Handle("/doc/", (*srvdoc)(mod))
}

type srvraw htmod

func (s *srvraw) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	res := s.ws.Res(ws.NewId(path))
	if res == nil || res.Flag&(ws.FlagDir|ws.FlagIgnore) != 0 {
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

type srvdoc htmod

func (s *srvdoc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[:5] != "/doc/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path[5:]
	pkg := s.src.Find(path)
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
