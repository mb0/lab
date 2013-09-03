// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package htmod

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/mb0/lab/ws"
)

func (mod *htmod) serveContent() {
	http.Handle("/raw/", (*srvraw)(mod))
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
