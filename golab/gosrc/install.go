// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var Gotool = filepath.Join(runtime.GOROOT(), "bin/go")

func Install(pkg *Pkg) *Report {
	r := gocmd("install", pkg.Path)
	if r.Start() == nil {
		r.WaitTimeout(time.Second)
	}
	return r
}

func Test(pkg *Pkg) *Report {
	r := gocmd("test", "-v", pkg.Path)
	if r.Start() == nil {
		r.WaitTimeout(time.Second)
	}
	return r
}

func gocmd(args ...string) *Report {
	r := &Report{}
	r.Command = &exec.Cmd{
		Path:   Gotool,
		Args:   append([]string{"go"}, args...),
		Dir:    os.TempDir(),
		Stdout: &r.Stdout,
		Stderr: &r.Stderr,
	}
	return r
}
