// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"bytes"
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

func init() {
	path := filepath.Join(runtime.GOROOT(), "bin/go")
	_, err := os.Stat(path)
	if err != nil {
		path, err = exec.LookPath("go")
		if err != nil {
			fmt.Println("could not find go tool")
			path = ""
		}
	}
	gotool = path
	if runtime.GOOS == "windows" {
		testexe += ".exe"
	}
}

type Result struct {
	Mode   string
	Time   int64
	Errmsg string `json:",omitempty"`
	Output []string
}

func Install(pkg *Pkg) *Result {
	r := &Result{Mode: "install"}
	cmd := newcmd(gotool, "go", "install", pkg.Path)

	err := cmd.Start()
	if err != nil {
		r.Errmsg = err.Error()
		return r
	}
	r.Time = time.Now().Unix()
	wait(cmd, r)
	return r
}

func Test(pkg *Pkg) *Result {
	r := &Result{Mode: "test"}
	cmd := newcmd(gotool, "go", "test", "-c", "-i", pkg.Path)

	tmp, err := ioutil.TempDir("", "labtest")
	if err != nil {
		r.Errmsg = err.Error()
		return r
	}
	cmd.Dir = tmp
	defer os.RemoveAll(tmp)

	err = cmd.Start()
	if err != nil {
		r.Errmsg = err.Error()
		return r
	}
	r.Time = time.Now().Unix()
	if wait(cmd, r); r.Errmsg != "" {
		return r
	}
	_, binary := filepath.Split(pkg.Path)
	binary += testexe
	cmd = newcmd(filepath.Join(tmp, binary), binary, "-test.v", "-test.short", "-test.timeout=3s")
	cmd.Dir = pkg.Dir

	err = cmd.Start()
	if err != nil {
		r.Errmsg = err.Error()
		return r
	}
	wait(cmd, r)
	return r
}

func newcmd(args ...string) *exec.Cmd {
	var buf bytes.Buffer
	return &exec.Cmd{
		Path:   args[0],
		Args:   args[1:],
		Dir:    os.TempDir(),
		Stdout: &buf,
		Stderr: &buf,
	}
}
func wait(cmd *exec.Cmd, r *Result) {
	err := cmd.Wait()
	if err != nil {
		r.Errmsg = err.Error()
	}
	b, l := line(cmd.Stdout.(*bytes.Buffer).Bytes())
	for ; len(l) > 0; b, l = line(b) {
		if len(l) > 1 {
			r.Output = append(r.Output, string(l))
		}
	}
}


func line(buf []byte) ([]byte, []byte) {
	if i := bytes.IndexByte(buf, '\n'); i > -1 {
		return buf[i+1:], buf[:i+1]
	}
	return nil, buf
}