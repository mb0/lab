// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package htmod

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	*ot.Server
	ws.Id
	Path  string
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
	if r.Flag&(ws.FlagIgnore|ws.FlagDir) != 0 {
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
	log.Println("file changed ot doc must be updated.")
}

func (mod *htmod) docroute(m hub.Msg, from hub.Id) {
	var rev apiRev
	to := from
	err := m.Unmarshal(&rev)
	if err != nil {
		log.Println(err)
		return
	}
	id := ws.Id(rev.Id)
	mod.docs.Lock()
	defer mod.docs.Unlock()
	doc, found := mod.docs.all[id]
	if !found {
		if m.Head != "subscribe" {
			log.Println("doc not found")
			return
		}
		r := mod.ws.Res(id)
		if r == nil {
			log.Println("res not found")
			return
		}
		if r.Flag&(ws.FlagIgnore|ws.FlagDir) != 0 {
			log.Println("ignored or dir")
			return
		}
		path := r.Path()
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println(err)
			return
		}
		doc = &otdoc{Id: id, Path: path, Server: &ot.Server{}}
		doc.Doc = (*ot.Doc)(&data)
		mod.docs.all[doc.Id] = doc
		mod.Hub.Add <- doc
	}
	doc.Lock()
	defer doc.Unlock()
	switch m.Head {
	case "subscribe":
		doc.group = append(doc.group, hub.Id(rev.User))
		m, err = hub.Marshal("subscribe", apiDoc{
			Id:  rev.Id,
			Rev: doc.Rev(),
			Doc: string(*doc.Doc),
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
		m, err = hub.Marshal("unsubscribe", apiRev{
			Id: id,
		})
	case "revise":
		ops, err := doc.Recv(rev.Rev, rev.Ops)
		if err != nil {
			log.Println(err)
			return
		}
		to = doc.GroupId()
		m, err = hub.Marshal("revise", apiRev{
			Id:   id,
			Rev:  doc.Rev(),
			Ops:  ops,
			User: from,
		})
	case "publish":
		// write to file
		var f *os.File
		f, err = os.OpenFile(doc.Path, os.O_WRONLY|os.O_TRUNC, 0)
		if err != nil {
			log.Println(err)
			return
		}
		var n int
		data := ([]byte)(*doc.Doc)
		n, err = f.Write(data)
		f.Close()
		if err != nil {
			log.Println(err)
			return
		}
		f.Close()
		if n < len(data) {
			log.Println("short write")
			return
		}
		to = doc.GroupId()
		m, err = hub.Marshal("publish", apiRev{
			Id:   id,
			Rev:  doc.Rev(),
			User: from,
		})
	}
	if err != nil {
		log.Println(err)
		return
	}
	mod.SendMsg(m, to)
}
