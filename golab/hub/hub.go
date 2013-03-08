// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hub

import (
	"encoding/json"
	"net/http"
)

type Id int64

const (
	Router Id = 0
	Group  Id = 1 << 32
)

type Msg struct {
	Head string
	Data *json.RawMessage `json:",omitempty"`
}

func Marshal(head string, v interface{}) (m Msg, err error) {
	var data []byte
	data, err = json.Marshal(v)
	if err != nil {
		return
	}
	m.Head = head
	m.Data = (*json.RawMessage)(&data)
	return
}

func (m *Msg) Unmarshal(v interface{}) error {
	return json.Unmarshal([]byte(*m.Data), v)
}

var (
	Signon  = Msg{Head: "_signon"}
	Signoff = Msg{Head: "_signon"}
)

type Envelope struct {
	From, To Id
	Msg
}

type Hub struct {
	conns  map[Id]*conn
	groups map[Id][]Id
	router func(hub *Hub, msg Msg, from Id)

	signon  chan *conn
	signoff chan *conn
	send    chan Envelope
}

func New(router func(*Hub, Msg, Id)) *Hub {
	h := &Hub{
		conns:  make(map[Id]*conn),
		groups: make(map[Id][]Id),
		router: router,

		signon:  make(chan *conn, 8),
		signoff: make(chan *conn, 8),
		send:    make(chan Envelope, 64),
	}
	go h.run()
	return h
}
func (h *Hub) run() {
	for {
		select {
		case c := <-h.signon:
			h.conns[c.id] = c
			h.router(h, Signon, c.id)
		case c := <-h.signoff:
			_, ok := h.conns[c.id]
			if !ok {
				continue
			}
			h.router(h, Signoff, c.id)
			delete(h.conns, c.id)
			if _, ok := <-c.send; ok {
				close(c.send)
			}
		case e := <-h.send:
			switch e.To {
			case Router:
				h.router(h, e.Msg, e.From)
			default:
				h.sendto(e.Msg, e.To)
			}
		}
	}
}
func (h *Hub) sendto(msg Msg, to Id) {
	if to&Group == 0 {
		if c, ok := h.conns[to]; ok {
			c.send <- msg
		}
		return
	}
	if to^Group == 0 {
		for _, c := range h.conns {
			c.send <- msg
		}
		return
	}
	for _, id := range h.groups[to] {
		h.sendto(msg, id)
	}
}

func (h *Hub) Send(e Envelope) {
	h.send <- e
}
func (h *Hub) SendMsg(m Msg, to Id) {
	h.send <- Envelope{Router, to, m}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	c, err := newconn(w, r)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	h.signon <- c
	go c.write()
	c.read(h)
	h.signoff <- c
}
