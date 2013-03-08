// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hub

import (
	"encoding/json"
	"github.com/garyburd/go-websocket/websocket"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	writeWait  = 10 * time.Second
	readWait   = 60 * time.Second
	pingPeriod = (readWait * 9) / 10
)

type conn struct {
	id     Id
	send   chan Msg
	wconn  *websocket.Conn
	ticker *time.Ticker
}

func newconn(w http.ResponseWriter, r *http.Request) (*conn, error) {
	wconn, err := websocket.Upgrade(w, r.Header, "", 1024, 1024)
	if err != nil {
		return nil, err
	}
	hash := fnv.New32()
	hash.Write([]byte(r.RemoteAddr))
	return &conn{Id(hash.Sum32()), make(chan Msg, 64), wconn, time.NewTicker(pingPeriod)}, nil
}

func (c *conn) read(h *Hub) {
	defer c.close()
	for {
		c.wconn.SetReadDeadline(time.Now().Add(readWait))
		op, r, err := c.wconn.NextReader()
		if err != nil {
			log.Println("error receiving message", err)
			return
		}
		if op == websocket.OpBinary {
			return
		}
		if op != websocket.OpText {
			continue
		}
		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			log.Println("error receiving message", err)
			return
		}
		var msg Msg
		err = json.Unmarshal(bytes, &msg)
		if err != nil {
			log.Println("error decoding message", err)
			return
		}
		h.router(h, msg, c.id)
	}
}

func (c *conn) write() {
	defer c.close()
	for {
		select {
		case msg, ok := <-c.send:
			c.wconn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.wconn.WriteMessage(websocket.OpClose, []byte{})
				return
			}
			w, err := c.wconn.NextWriter(websocket.OpText)
			if err != nil {
				return
			}
			enc := json.NewEncoder(w)
			err = enc.Encode(msg)
			if err != nil {
				log.Println("error encoding message", err)
			}
			err = w.Close()
			if err != nil {
				log.Println("error sending message", err)
				return
			}
		case now := <-c.ticker.C:
			c.wconn.SetWriteDeadline(now.Add(writeWait))
			err := c.wconn.WriteMessage(websocket.OpPing, []byte{})
			if err != nil {
				log.Println("error sending ping", err)
				return
			}
		}
	}
}

func (c *conn) close() {
	c.ticker.Stop()
	c.wconn.Close()
}
