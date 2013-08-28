// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	_ "net/http/pprof"

	_ "github.com/mb0/ace"
	"github.com/mb0/lab"
	"github.com/mb0/lab/golab/htmod"
)

var (
	htaddr     = lab.Conf.String("addr", "localhost:8910", "http server addr")
	useHttp    = lab.Conf.Bool("http", false, "start http server")
	useHttps   = lab.Conf.Bool("https", false, "start https server")
	keyFile    = lab.Conf.String("key", "", "key file  for ssl")
	certFile   = lab.Conf.String("cert", "", "cert file for ssl")
	cacertFile = lab.Conf.String("cacert", "", "client ca cert file for authentication")
)

func init() {
	lab.LoadConf()
	if !(*useHttp || *useHttps) {
		return
	}
	conf := htmod.Config{
		Https: *useHttps,
		Addr:  *htaddr,
	}
	if conf.Https {
		conf.KeyFile = *keyFile
		conf.CertFile = *certFile
		conf.CAFile = *cacertFile
		log.Printf("starting https://%s/\n", conf.Addr)
	} else {
		log.Printf("starting http://%s/\n", conf.Addr)
	}
	lab.Register("htmod", htmod.New(conf))
}
