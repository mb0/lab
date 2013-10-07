// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package htmod

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/mb0/lab"
	"github.com/mb0/lab/golab/gosrc"
	"github.com/mb0/lab/hub"
	"github.com/mb0/lab/ws"
	"github.com/mb0/lab/ws/find"
)

// BUG: https://code.google.com/p/go/issues/detail?id=6121
// use all non ECDHE-ECDSA ciphers
var ciphers = []uint16{
	tls.TLS_RSA_WITH_RC4_128_SHA,
	tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
}

type htmod struct {
	conf  Config
	roots []string
	ws    *ws.Ws
	src   *gosrc.Src
	docs  *docs
	*hub.Hub
}

type Config struct {
	Https    bool
	Addr     string
	KeyFile  string
	CertFile string
	CAFile   string
}

func New(conf Config) *htmod {
	return &htmod{conf: conf}
}

func (mod *htmod) Init() {
	mod.roots = lab.Mod("roots").([]string)
	mod.ws = lab.Mod("ws").(*ws.Ws)
	mod.src = lab.Mod("gosrc").(*gosrc.Src)
	mod.serveStatic()
}

func (mod *htmod) Run() {
	mod.Hub = hub.New()
	mod.docs = &docs{all: make(map[ws.Id]*otdoc)}
	go func() {
		for e := range mod.Hub.Route {
			mod.route(e.Msg, e.From)
		}
	}()
	mod.src.SignalReports(func(r *gosrc.Report) {
		m, err := hub.Marshal("report", []*gosrc.Report{r})
		if err != nil {
			log.Println(err)
			return
		}
		mod.SendMsg(m, hub.Group)
	})
	http.Handle("/ws", mod.Hub)
	var err error
	server := &http.Server{
		Addr: mod.conf.Addr,
	}
	if mod.conf.Https {
		if mod.conf.CAFile != "" {
			pemByte, err := ioutil.ReadFile(mod.conf.CAFile)
			if err != nil {
				log.Fatalf("reading ca file:\n\t%s\n", err)
			}
			block, _ := pem.Decode(pemByte)
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				log.Fatalf("parsing ca file:\n\t%s\n", err)
			}
			pool := x509.NewCertPool()
			pool.AddCert(cert)
			server.TLSConfig = &tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    pool,
				CipherSuites: ciphers,
			}
		}
		err = server.ListenAndServeTLS(mod.conf.CertFile, mod.conf.KeyFile)
	} else {
		err = server.ListenAndServe()
	}
	if err != nil {
		log.Fatalf("http %s\n", err)
	}
}

func (mod *htmod) route(m hub.Msg, id hub.Id) {
	var (
		err error
		msg hub.Msg
	)
	switch m.Head {
	case hub.Signon:
		// send reports for all working packages
		msg, err = hub.Marshal("report", mod.src.AllReports())
	case "stat":
		var path string
		if err = m.Unmarshal(&path); err != nil {
			break
		}
		msg, err = mod.stat(path)
	case "find":
		var query string
		if err = m.Unmarshal(&query); err != nil {
			break
		}
		msg, err = mod.find(query)
	case "subscribe", "unsubscribe", "revise", "publish":
		mod.docroute(m, id)
		return
	case "complete", "format":
		mod.actionRoute(m, id)
		return
	default:
		msg, err = hub.Marshal("unknown", m.Head)
	}
	if err != nil {
		log.Println(err)
		return
	}
	mod.SendMsg(msg, id)
}

func (mod *htmod) stat(path string) (hub.Msg, error) {
	res := apiRes{ws.NewId(path), path, false, path}
	if r := mod.ws.Res(res.Id); r != nil {
		r.Lock()
		defer r.Unlock()
		res.Name, res.IsDir = r.Name, r.Flag&ws.FlagDir != 0
		if r.Dir != nil {
			cs := make([]apiRes, 0, len(r.Children))
			for _, c := range r.Children {
				if c.Flag&ws.FlagIgnore == 0 {
					cs = append(cs, apiRes{
						c.Id,
						c.Name,
						c.Flag&ws.FlagDir != 0,
						"",
					})
				}
			}
			return hub.Marshal("stat", apiStat{res, cs, ""})
		}
		return hub.Marshal("stat", apiStat{res, nil, ""})
	}
	return hub.Marshal("stat", apiStat{res, nil, "not found"})
}

type apiStat struct {
	apiRes
	Children []apiRes `json:",omitempty"`
	Error    string   `json:",omitempty"`
}

func (mod *htmod) find(query string) (hub.Msg, error) {
	mod.ws.Lock()
	defer mod.ws.Unlock()
	list, err := find.Find(mod.ws, query)
	if err != nil {
		return hub.Marshal("find", apiFind{query, nil, err.Error()})
	}
	res := make([]apiRes, 0, len(list))
	for _, r := range list {
		res = append(res, apiRes{
			r.Id,
			r.Name,
			r.Flag&ws.FlagDir != 0,
			r.Path(),
		})
	}
	return hub.Marshal("find", apiFind{query, res, ""})
}

type apiFind struct {
	Query  string
	Result []apiRes `json:",omitempty"`
	Error  string   `json:",omitempty"`
}

type apiRes struct {
	Id    ws.Id
	Name  string
	IsDir bool
	Path  string `json:",omitempty"`
}
