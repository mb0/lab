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
	mod.Hub = hub.New(func(h *hub.Hub, m hub.Msg, id hub.Id) {
		switch m.Head {
		case hub.Signon.Head:
			// send reports for all working packages
			msg, err := hub.Marshal("reports", mod.src.AllReports())
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
			msg, err := mod.apistat(path)
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

func (mod *htmod) apistat(path string) (hub.Msg, error) {
	res := apires{ws.NewId(path), path, false}
	if r := mod.ws.Res(res.Id); r != nil {
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
