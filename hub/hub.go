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
	Route Id = 0
	Group Id = 1 << 32
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

	signon  chan *conn
	signoff chan *conn
	route   chan Envelope
	send    chan Envelope
}

func New() *Hub {
	h := &Hub{
		conns:  make(map[Id]*conn),
		groups: make(map[Id][]Id),

		signon:  make(chan *conn, 8),
		signoff: make(chan *conn, 8),
		route:   make(chan Envelope, 64),
		send:    make(chan Envelope, 64),
	}
	go h.run()
	return h
}
func (h *Hub) Route() <-chan Envelope {
	return h.route
}
func (h *Hub) run() {
	for {
		select {
		case c := <-h.signon:
			h.conns[c.id] = c
			h.route <- Envelope{c.id, Route, Signon}
		case c := <-h.signoff:
			delete(h.conns, c.id)
			c.close()
			close(c.send)
			h.route <- Envelope{c.id, Route, Signoff}
		case e := <-h.send:
			switch e.To {
			case Route:
				h.route <- e
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
	h.send <- Envelope{Route, to, m}
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
