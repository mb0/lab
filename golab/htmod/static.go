// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package htmod

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mb0/ace"
)

func (mod *htmod) serveStatic() {
	ng := mod.findsrc("github.com/mb0/lab/goapp")
	ace := mod.findsrc("github.com/mb0/ace/lib/ace")
	if ng == "" || ace == "" {
		log.Fatalf("cannot find client files.\n")
	}
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(ng))))
	http.Handle("/ace/", http.StripPrefix("/ace/", http.FileServer(http.Dir(ace))))
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
