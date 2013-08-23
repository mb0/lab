// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	_ "net/http/pprof"

	_ "github.com/mb0/ace"
	"github.com/mb0/lab"
	"github.com/mb0/lab/golab/htmod"
)

var (
	htaddr     = flag.String("addr", "localhost:8910", "http server addr")
	useHttp    = flag.Bool("http", false, "start http server")
	useHttps   = flag.Bool("https", false, "start https server")
	keyFile    = flag.String("key", "", "key file  for ssl")
	certFile   = flag.String("cert", "", "cert file for ssl")
	cacertFile = flag.String("cacert", "", "client ca cert file for authentication")
)

func init() {
	flag.Parse()
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
