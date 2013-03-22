// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package htmod

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mb0/lab/hub"
	"github.com/mb0/lab/ws"
)

var (
	gofmt  string
	gocode string
)

func exists(dir, rel string) string {
	cmd := filepath.Join(dir, rel)
	if _, err := os.Stat(cmd); err == nil {
		return cmd
	}
	return ""
}

func init() {
	root := runtime.GOROOT()
	list := filepath.SplitList(build.Default.GOPATH)
	gofmt = exists(root, "bin/gofmt")
	for _, p := range append(list, root) {
		if gocode = exists(p, "bin/gocode"); gocode != "" {
			break
		}
	}
	log.Println("found tools:", gofmt, gocode)
}

type actionReq struct {
	Id   ws.Id
	Offs int
}

func (mod *htmod) actionRoute(m hub.Msg, from hub.Id) {
	var req actionReq
	err := m.Unmarshal(&req)
	if err != nil {
		log.Println(err)
		return
	}
	mod.docs.RLock()
	doc, found := mod.docs.all[req.Id]
	mod.docs.RUnlock()
	if !found {
		log.Println("no document open with id", req)
		return
	}
	if pl := len(doc.Path); pl < 3 || doc.Path[pl-3:] != ".go" {
		log.Println("document not a go file", doc.Path)
		return
	}
	var cmd *exec.Cmd
	switch {
	case m.Head == "complete" && gocode != "":
		cmd = &exec.Cmd{
			Path: gocode,
			Args: []string{
				"gocode",
				"-f=json",
				"autocomplete",
				fmt.Sprint(req.Offs),
			},
		}
	case m.Head == "format" && gofmt != "":
		cmd = &exec.Cmd{
			Path: gofmt,
			Args: []string{
				"gofmt",
			},
		}
	default:
		log.Println("unexpected msg", m.Head)
		return
	}
	var buf bytes.Buffer
	doc.Lock()
	buf.Write(([]byte)(*doc.Doc))
	doc.Unlock()
	cmd.Stdin = &buf
	data, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err, data)
		return
	}
	switch {
	case m.Head == "complete":
		if len(data) < 10 {
			return
		}
		data = data[4:len(data)-1]
		m, err = hub.Marshal("complete", struct {
			actionReq
			Proposed *json.RawMessage
		}{req, (*json.RawMessage)(&data)})
		if err != nil {
			log.Println(err)
			return
		}
		mod.SendMsg(m, from)
	case m.Head == "format":
		doc.Lock()
		defer doc.Unlock()
		rev := doc.Rev()
		ops := doc.diffops(data)
		if ops != nil {
			mod.handlerev("revise", apiRev{Id: req.Id, Rev: rev, Ops: ops}, doc)
		}
	}
}
