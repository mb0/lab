// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ot

import (
	"testing"
)

func TestDocApply(t *testing.T) {
	doc := Doc("abc")
	if err := doc.Apply(Ops{{N: 1}, {S: "tag"}, {N: -2}}); err != nil {
		t.Error(err)
	}
	if got := string(doc); got != "atag" {
		t.Errorf("expected atag got %s", got)
	}
}

func TestServer(t *testing.T) {
	doc := Doc("abc")
	s := &Server{&doc, nil}
	_, rev, err := s.Recv(1, Ops{})
	if err == nil || rev != 0 {
		t.Error("expected error")
	}
	a := Ops{{N: 1}, {S: "tag"}, {N: 2}}
	a1, rev, err := s.Recv(0, a)
	if err != nil || rev != 1 {
		t.Error(err)
	}
	if !a1.Equal(a) {
		t.Errorf("expected %v got %v", a, a1)
	}
	b1, rev, err := s.Recv(0, Ops{{N: 1}, {N: -2}})
	if err != nil || rev != 2 {
		t.Error(err)
	}
	if !b1.Equal(Ops{{N: 4}, {N: -2}}) {
		t.Errorf("expected %v got %v", a, a1)
	}
}

func TestClient(t *testing.T) {
	var sent []Ops
	doc := Doc("old!")
	c := &Client{Doc: &doc, Send: func(rev int, ops Ops) {
		sent = append(sent, ops)
	}}
	a := Ops{{S: "g"}, {N: 4}}
	err := c.Apply(a)
	if err != nil {
		t.Error(err)
	}
	if s := string(doc); s != "gold!" {
		t.Errorf(`expected  "gold!" got %q`, s)
	}
	if !a.Equal(c.Wait) || !a.Equal(sent[0]) {
		t.Error("expected waiting for ack")
	}
	b := Ops{{N: 2}, {N: -2}, {N: 1}}
	err = c.Apply(b)
	if err != nil {
		t.Error(err)
	}
	if s := string(doc); s != "go!" {
		t.Errorf(`expected  "go!" got %q`, s)
	}
	if !b.Equal(c.Buf) || len(sent) != 1 {
		t.Error("expected buffering")
	}
	err = c.Apply(Ops{{N: 2}, {S: " cool"}, {N: 1}})
	if err != nil {
		t.Error(err)
	}
	if s := string(doc); s != "go cool!" {
		t.Errorf(`expected  "go cool!" got %q`, s)
	}
	cb := Ops{{N: 2}, {N: -2}, {S: " cool"}, {N: 1}}
	if !cb.Equal(c.Buf) || len(sent) != 1 {
		t.Error("expected combinig buffer")
	}
	err = c.Recv(Ops{{N: 1}, {S: " is"}, {N: 3}})
	if err != nil {
		t.Error(err)
	}
	if s := string(doc); s != "go is cool!" {
		t.Errorf(`expected  "go is cool!" got %q`, s)
	}
	if !c.Wait.Equal(Ops{{S: "g"}, {N: 7}}) {
		t.Error("expected transform wait", c.Wait)
	}
	cb = Ops{{N: 5}, {N: -2}, {S: " cool"}, {N: 1}}
	if !c.Buf.Equal(cb) {
		t.Error("expected transform buf", c.Buf)
	}
	err = c.Ack()
	if err != nil {
		t.Error(err)
	}
	if c.Buf != nil || !cb.Equal(c.Wait) || len(sent) != 2 || !cb.Equal(sent[1]) {
		t.Error("expected flushing buffer")
	}
	err = c.Ack()
	if err != nil {
		t.Error(err)
	}
	if c.Buf != nil || c.Wait != nil || len(sent) != 2 {
		t.Error("expected flushed")
	}
}
