// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package ws

// the walker code is heavily inspired by package github.com/kr/fs

type Walker struct {
	cur   *Res
	stack []*Res
	desc  bool
}

// Walk returns a walker that visits the descendants of r.
func Walk(r *Res) *Walker {
	return &Walker{r, nil, true}
}

// WalkAll returns a walker that visits all workspace resources.
func WalkAll(ws *Ws) *Walker {
	return &Walker{ws.root, nil, true}
}

// Res returns the current resource.
func (w *Walker) Res() *Res {
	return w.cur
}

// Skip prohibits descension into the current resource.
func (w *Walker) Skip() {
	w.desc = false
}

// Next advances to the next resource or returns false if at the end of the tree.
func (w *Walker) Next() bool {
	if w.desc {
		w.cur.Lock()
		if w.cur.Dir != nil {
			list := w.cur.Children
			for i := len(list) - 1; i >= 0; i-- {
				w.stack = append(w.stack, list[i])
			}
		}
		w.cur.Unlock()
	}
	if len(w.stack) == 0 {
		return false
	}
	i := len(w.stack) - 1
	w.cur = w.stack[i]
	w.stack = w.stack[:i]
	w.desc = true
	return true
}
