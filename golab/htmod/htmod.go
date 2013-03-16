// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package htmod

import (
	"log"
	"net/http"

	"github.com/mb0/lab"
	"github.com/mb0/lab/golab/gosrc"
	"github.com/mb0/lab/hub"
	"github.com/mb0/lab/ws"
)

type htmod struct {
	addr  string
	roots []string
	ws    *ws.Ws
	src   *gosrc.Src
	docs  *docs
	*hub.Hub
}

func New(addr string) *htmod {
	return &htmod{addr: addr}
}

func (mod *htmod) Init() {
	mod.roots = lab.Mod("roots").([]string)
	mod.ws = lab.Mod("ws").(*ws.Ws)
	mod.src = lab.Mod("gosrc").(*gosrc.Src)
	mod.serveStatic()
	mod.serveContent()
}

func (mod *htmod) Run() {
	mod.Hub = hub.New()
	mod.docs = &docs{all: make(map[ws.Id]*otdoc)}
	go func() {
		for e := range mod.Hub.Route {
			mod.route(e.Msg, e.From)
		}
	}()
	mod.src.SignalReports(func(r *gosrc.Report) {
		m, err := hub.Marshal("report", r)
		if err != nil {
			log.Println(err)
			return
		}
		mod.SendMsg(m, hub.Group)
	})
	http.Handle("/ws", mod.Hub)
	err := http.ListenAndServe(mod.addr, nil)
	if err != nil {
		log.Fatalf("http %s\n", err)
	}
}

func (mod *htmod) route(m hub.Msg, id hub.Id) {
	var (
		err error
		msg hub.Msg
	)
	switch m.Head {
	case hub.Signon:
		// send reports for all working packages
		msg, err = hub.Marshal("reports", mod.src.AllReports())
	case "stat":
		var path string
		if err = m.Unmarshal(&path); err != nil {
			break
		}
		msg, err = mod.stat(path)
	case "subscribe", "unsubscribe", "revise", "publish":
		mod.docroute(m, id)
		return
	default:
		msg, err = hub.Marshal("unknown", m.Head)
	}
	if err != nil {
		log.Println(err)
		return
	}
	mod.SendMsg(msg, id)
}

func (mod *htmod) stat(path string) (hub.Msg, error) {
	res := apiRes{ws.NewId(path), path, false}
	if r := mod.ws.Res(res.Id); r != nil {
		r.Lock()
		defer r.Unlock()
		res.Name, res.IsDir = r.Name, r.Flag&ws.FlagDir != 0
		if r.Dir != nil {
			cs := make([]apiRes, 0, len(r.Children))
			for _, c := range r.Children {
				if c.Flag&ws.FlagIgnore == 0 {
					cs = append(cs, apiRes{
						c.Id,
						c.Name,
						c.Flag&ws.FlagDir != 0,
					})
				}
			}
			return hub.Marshal("stat", struct {
				apiRes
				Path     string
				Children []apiRes
			}{res, path, cs})
		}
		return hub.Marshal("stat", struct {
			apiRes
			Path string
		}{res, path})
	}
	return hub.Marshal("stat.err", struct {
		apiRes
		Path  string
		Error string
	}{res, path, "not found"})
}

type apiRes struct {
	Id    ws.Id
	Name  string
	IsDir bool
}
