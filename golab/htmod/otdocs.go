// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package htmod

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/mb0/lab/hub"
	"github.com/mb0/lab/ot"
	"github.com/mb0/lab/ws"
)

var DocGroup hub.Id = (1 << 40) | hub.Group

type rev struct {
	Rev int
	Doc string
}

type otdoc struct {
	sync.Mutex
	ws.Id
	*ot.Server
	group []hub.Id
}

func (doc *otdoc) GroupId() hub.Id {
	return hub.Id(doc.Id) | DocGroup
}
func (doc *otdoc) Group() []hub.Id {
	doc.Lock()
	defer doc.Unlock()
	group := make([]hub.Id, len(doc.group))
	copy(group, doc.group)
	return group
}

type docs struct {
	sync.RWMutex
	all map[ws.Id]*otdoc
}

type apiDoc struct {
	Id  ws.Id
	Rev int
	Doc string
}
type apiRev struct {
	Id   ws.Id
	Rev  int
	Ops  ot.Ops `json:",omitempty"`
	User hub.Id
}

func (mod *htmod) Handle(op ws.Op, r *ws.Res) {
	if op&(ws.Modify|ws.Delete) == 0 {
		return
	}
	mod.docs.RLock()
	_, found := mod.docs.all[r.Id]
	mod.docs.RUnlock()
	if !found {
		return
	}
	mod.docs.Lock()
	defer mod.docs.Unlock()
	doc := mod.docs.all[r.Id]
	if doc == nil {
		return
	}
	doc.Lock()
	defer doc.Unlock()
	// delete doc
	if op&ws.Delete != 0 {
		mod.Hub.Del <- doc
		delete(mod.docs.all, doc.Id)
		msg, err := hub.Marshal("unsubscribe", apiRev{
			Id:   r.Id,
			User: DocGroup,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, cid := range doc.group {
			mod.Hub.SendMsg(msg, cid)
		}
		return
	}
	// update from filesystem
	data, err := ioutil.ReadFile(r.Path())
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO diff data and broadcast ops
	_ = data
}

func (mod *htmod) docroute(m hub.Msg, from hub.Id) (hub.Msg, error) {
	var rev apiRev
	err := m.Unmarshal(&rev)
	if err != nil {
		return m, err
	}
	id := ws.Id(rev.Id)
	mod.docs.Lock()
	defer mod.docs.Unlock()
	doc, found := mod.docs.all[id]
	if !found {
		if m.Head != "subscribe" {
			return m, fmt.Errorf("res not found")
		}
		res := mod.ws.Res(id)
		if res == nil {
			return m, fmt.Errorf("res not found")
		}
		data, err := ioutil.ReadFile(res.Path())
		if err != nil {
			return m, err
		}
		doc = &otdoc{Id: id, Server: &ot.Server{}}
		doc.Doc = (*ot.Doc)(&data)
		mod.docs.all[doc.Id] = doc
		mod.Hub.Add <- doc
	}
	doc.Lock()
	defer doc.Unlock()
	switch m.Head {
	case "subscribe":
		doc.group = append(doc.group, hub.Id(rev.User))
		return hub.Marshal("subscribe", apiDoc{
			Id:  rev.Id,
			Rev: doc.Rev(),
			Doc: string(*doc.Doc),
		})
	case "revise":
		ops, err := doc.Recv(rev.Rev, rev.Ops)
		if err != nil {
			return m, err
		}
		return hub.Marshal("revise", apiRev{
			Id:   id,
			Rev:  doc.Rev(),
			Ops:  ops,
			User: from,
		})
	case "unsubscribe":
		for i, cid := range doc.group {
			if cid != rev.User {
				continue
			}
			doc.group = append(doc.group[:i], doc.group[i+1:]...)
			if len(doc.group) == 0 {
				mod.Hub.Del <- doc
			}
			break
		}
		return hub.Marshal("unsubscribe", apiRev{
			Id:   id,
			User: from,
		})
	case "publish":
		// TODO write to file
	}
	return hub.Marshal("unknown", m.Head)
}
