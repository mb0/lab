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
	htstart = flag.Bool("http", false, "start http server")
	htaddr  = flag.String("addr", "localhost:8910", "http server addr")
)

func init() {
	flag.Parse()
	if *htstart {
		log.Printf("starting http://%s/\n", *htaddr)
		lab.Register("htmod", htmod.New(*htaddr))
	}
}
