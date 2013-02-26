// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

var (
	failmsg = "\x1b[41mFAIL\x1b[0m"
	failpre = "\x1b[41m    \x1b[0m"
	okmsg   = "\x1b[42mok  \x1b[0m"
)

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

func (r *Report) String() string {
	if r.Command == nil || len(r.Command.Args) < 2 {
		return "<empty report>"
	}
	mode := r.Command.Args[1]
	path := r.Command.Args[len(r.Command.Args)-1]
	if r.Error != nil {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%s %-7s %s %s", failmsg, mode, path, r.Error)
		for _, b := range []*bytes.Buffer{&r.Stdout, &r.Stderr} {
			for l := line(b); l != nil; l = line(b) {
				if len(l) == 0 || l[0] == '#' {
					continue
				}
				fmt.Fprintf(&buf, "\n%s ", failpre)
				buf.Write(l)
			}
		}
		return buf.String()
	}
	return fmt.Sprintf("%s %-7s %s", okmsg, mode, path)
}

func line(buf *bytes.Buffer) []byte {
	line, err := buf.ReadBytes('\n')
	if err != nil {
		return line
	}
	if len(line) < 1 {
		return nil
	}
	return line[:len(line)-1]
}
