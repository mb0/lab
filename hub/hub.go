// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hub

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Id int64

func (id *Id) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`"%X"`, id)
	return []byte(str), nil
}
func (id *Id) UnmarshalJSON(data []byte) error {
	_, err := fmt.Sscanf(string(data), `"%X"`, id)
	return err
}

const (
	Route  Id = 0
	Group  Id = 1 << 32
	Except Id = 2 << 32
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
	Signon  = "_signon"
	Signoff = "_signoff"
)

type Envelope struct {
	From, To Id
	Msg
}

type Grouper interface {
	GroupId() Id
	Group() []Id
}

type Hub struct {
	conns   map[Id]*conn
	groups  map[Id]Grouper
	signon  chan *conn
	signoff chan *conn
	Add     chan Grouper
	Del     chan Grouper
	Route   chan Envelope
	Send    chan Envelope
}

func New() *Hub {
	h := &Hub{
		conns:   make(map[Id]*conn),
		groups:  make(map[Id]Grouper),
		signon:  make(chan *conn, 8),
		signoff: make(chan *conn, 8),
		Add:     make(chan Grouper, 8),
		Del:     make(chan Grouper, 8),
		Route:   make(chan Envelope, 64),
		Send:    make(chan Envelope, 64),
	}
	go h.run()
	return h
}
func (h *Hub) run() {
	for {
		select {
		case c := <-h.signon:
			h.conns[c.id] = c
			h.Route <- Envelope{c.id, Route, Msg{Head: Signon}}
		case c := <-h.signoff:
			delete(h.conns, c.id)
			c.close()
			close(c.send)
			h.Route <- Envelope{c.id, Route, Msg{Head: Signoff}}
		case g := <-h.Add:
			h.groups[g.GroupId()] = g
		case g := <-h.Del:
			delete(h.groups, g.GroupId())
		case e := <-h.Send:
			h.send(e)
		}
	}
}

func (h *Hub) send(e Envelope) {
	var except Id
	if e.To&Except != 0 {
		e.To ^= Except
		except = e.From
	}
	switch {
	case e.To == Route:
		h.Route <- e
	case e.To == Group:
		for _, c := range h.conns {
			if c.id != except {
				c.send <- e.Msg
			}
		}
	case e.To&Group != 0:
		if g, ok := h.groups[e.To]; ok {
			for _, to := range g.Group() {
				if to == except {
					continue
				}
				if c, ok := h.conns[to]; ok {
					c.send <- e.Msg
				}
			}
		}
	default:
		if c, ok := h.conns[e.To]; ok {
			c.send <- e.Msg
		}
	}
}

func (h *Hub) SendMsg(m Msg, to Id) {
	h.Send <- Envelope{Route, to, m}
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
	select {
	case h.signon <- c:
		go c.write()
		c.read(h)
		h.signoff <- c
	default:
		http.Error(w, "Closing", http.StatusServiceUnavailable)
	}
}
