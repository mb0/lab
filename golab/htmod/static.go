// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package htmod

import (
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mb0/ace"
)

func (mod *htmod) serveStatic() {
	http.HandleFunc("/", index)
	static := mod.findsrc("github.com/mb0/lab/golab/static")
	statng := mod.findsrc("github.com/mb0/lab/goapp")
	statace := mod.findsrc("github.com/mb0/ace/lib/ace")
	if static == "" || statace == "" || statng == "" {
		indexbytes = []byte("cannot find client files.")
		return
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(static))))
	http.Handle("/ng/", http.StripPrefix("/ng/", http.FileServer(http.Dir(statng))))
	http.Handle("/static/ace/", http.StripPrefix("/static/ace/", http.FileServer(http.Dir(statace))))
	http.HandleFunc("/manifest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/cache-manifest; charset=utf-8")
		http.ServeFile(w, r, filepath.Join(static, "main.manifest"))
	})
}

func (mod *htmod) findsrc(path string) string {
	for _, dir := range mod.roots {
		p := filepath.Join(dir, path)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
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
	<script data-main="/static/main" src="//cdnjs.cloudflare.com/ajax/libs/require.js/2.1.4/require.min.js"></script>
</body></html>
`)
