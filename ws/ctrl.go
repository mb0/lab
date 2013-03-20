// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

// ctrl implements a workspace controller.
type ctrl Ws

func (w *ctrl) Control(op Op, id Id, name string) error {
	var r, p *Res
	w.Lock()
	defer w.Unlock()
	r = w.all[id]
	if name != "" {
		p, r = r, nil
		if p != nil && p.Dir != nil {
			p.Lock()
			r = find(p.Children, name)
			p.Unlock()
		}
	}
	switch {
	case op&Delete != 0:
		if r == nil {
			break
		}
		return w.remove(op, r)
	case r != nil:
		// res found, modify
		return w.change(op, r)
	case p != nil:
		// parent found create child
		return w.add(op, p, name)
	}
	// not found, ignore
	return nil
}
func (w *ctrl) change(fsop Op, r *Res) error {
	w.config.handle(fsop|Change, r)
	return nil
}
func (w *ctrl) remove(fsop Op, r *Res) error {
	r.Lock()
	if p := r.Parent; p != nil {
		p.Lock()
		defer p.Unlock()
		if p.Dir != nil {
			p.Children = remove(p.Children, r)
		}
	}
	r.Unlock()
	rm := []*Res{r}
	if r.Dir != nil {
		walk(r.Children, func(c *Res) error {
			rm = append(rm, c)
			return nil
		})
	}
	for _, c := range rm {
		c.Lock()
	}
	for i := len(rm) - 1; i >= 0; i-- {
		c := rm[i]
		w.config.handle(fsop|Remove, c)
		if c.Dir != nil {
			c.Children = nil
		}
		delete(w.all, c.Id)
		c.Unlock()
	}
	return nil
}
func (w *ctrl) add(fsop Op, p *Res, name string) error {
	p.Lock()
	defer p.Unlock()
	// new lock try again
	if find(p.Children, name) != nil {
		// ignore duplicate
		return nil
	}
	r, err := newChild(p, name, false, true)
	if err != nil {
		return err
	}
	p.Children = insert(p.Children, r)
	w.all[r.Id] = r
	switch {
	case w.config.filter(r):
		r.Flag |= FlagIgnore
		fallthrough
	case r.Dir == nil:
		w.config.handle(fsop|Add, r)
		return nil
	}
	if err = read(r, w.config.Filter); err != nil {
		return err
	}
	w.config.handle(fsop|Add, r)
	(*Ws)(w).addAllChildren(fsop, r)
	return nil
}
