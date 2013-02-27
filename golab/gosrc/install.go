// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var gotool = filepath.Join(runtime.GOROOT(), "bin/go")
var testexe = ".test"
var ErrTimeout = fmt.Errorf("timeout")

func init() {
	if runtime.GOOS == "windows" {
		testexe += ".exe"
	}
}

func Install(pkg *Pkg) *Report {
	r := &Report{Mode: "install", Path: pkg.Path}
	cmd := rcmd(r, gotool, "go", "install", r.Path)

	if r.Err = cmd.Start(); r.Err != nil {
		return r
	}
	r.Start = time.Now()
	r.Err = cmd.Wait()
	return r
}

func Test(pkg *Pkg) *Report {
	r := &Report{Mode: "test", Path: pkg.Path}
	cmd := rcmd(r, gotool, "go", "test", "-c", "-i", r.Path)

	tmp, err := ioutil.TempDir("", "labtest")
	if err != nil {
		r.Err = err
		return r
	}
	cmd.Dir = tmp
	defer os.RemoveAll(tmp)

	if r.Err = cmd.Start(); r.Err != nil {
		return r
	}
	r.Start = time.Now()
	r.Err = cmd.Wait()
	if r.Err != nil {
		return r
	}
	_, binary := filepath.Split(r.Path)
	binary += testexe
	cmd = rcmd(r, filepath.Join(tmp, binary), binary, "-test.v", "-test.short", "-test.timeout=3s")
	cmd.Dir = pkg.Dir
	if r.Err = cmd.Start(); r.Err != nil {
		return r
	}
	r.Err = cmd.Wait()
	return r
}

func rcmd(r *Report, args ...string) *exec.Cmd {
	return &exec.Cmd{
		Path:   args[0],
		Args:   args[1:],
		Dir:    os.TempDir(),
		Stdout: &r.out,
		Stderr: &r.err,
	}
}
