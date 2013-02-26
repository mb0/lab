// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var Gotool = filepath.Join(runtime.GOROOT(), "bin/go")

type Report struct {
	Command *exec.Cmd
	Started time.Time
	Ended   time.Time
	Error   error
	Stdout  bytes.Buffer
	Stderr  bytes.Buffer
}

func (r *Report) Start() error {
	r.Started = time.Now()
	r.Error = r.Command.Start()
	return r.Error
}
func (r *Report) Wait() error {
	if r.Error != nil {
		return r.Error
	}
	r.Error = r.Command.Wait()
	r.Ended = time.Now()
	return r.Error
}
func (r *Report) WaitTimeout(after time.Duration) error {
	if r.Error != nil {
		return r.Error
	}
	done := make(chan error)
	go func() {
		done <- r.Command.Wait()
	}()
	select {
	case err := <-done:
		r.Ended = time.Now()
		r.Error = err
	case <-time.After(after):
		r.Command.Process.Kill()
		r.Error = fmt.Errorf("timeout")
	}
	return r.Error
}

var (
	failmsg = "\x1b[41mFAIL\x1b[0m"
	okmsg   = "\x1b[42mok  \x1b[0m"
)

func (r *Report) String() string {
	if r.Command == nil || len(r.Command.Args) < 2 {
		return "<empty report>"
	}
	mode := r.Command.Args[1]
	path := r.Command.Args[len(r.Command.Args)-1]
	if r.Error != nil {
		return fmt.Sprintf("%s %-7s %s %s", failmsg, mode, path, r.Error)
	}
	return fmt.Sprintf("%s %-7s %s", okmsg, mode, path)
}

func Install(pkg *Pkg) *Report {
	r := gocmd("install", pkg.Path)
	r.Start()
	if r.Error == nil {
		r.WaitTimeout(time.Second)
	}
	return r
}

func Test(pkg *Pkg) *Report {
	r := gocmd("test", "-v", pkg.Path)
	r.Start()
	if r.Error == nil {
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
