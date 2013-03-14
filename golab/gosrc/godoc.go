// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var godoctool = filepath.Join(runtime.GOROOT(), "bin/godoc")
var godoctmpl string

func init() {
	tmpl := findTemplate()
	if tmpl != "" {
		godoctmpl = "-templates=" + tmpl
	}
}

var snips = [][]byte{
	[]byte(`<div id="nav">`),
	[]byte(`<div id="footer">`),
}

func LoadHtmlDoc(path string, all bool) ([]byte, error) {
	patharg := path
	if all {
		patharg += "?m=all"
	}
	var buf bytes.Buffer
	err := rungodoc(&buf, "godoc", fmt.Sprintf("-url=/pkg/%s", patharg), godoctmpl)
	if err != nil {
		return nil, err
	}
	return snip(buf.Bytes()), nil
}

func rungodoc(w io.Writer, args ...string) error {
	cmd := &exec.Cmd{
		Path:   godoctool,
		Args:   args,
		Stdout: w,
		Stderr: os.Stderr,
	}
	return cmd.Run()
}

func snip(in []byte) []byte {
	raw := in
	// snip header
	if i := bytes.Index(raw, snips[0]); i > -1 {
		if j := bytes.IndexByte(raw[i:], '\n'); j > -1 {
			raw = raw[i+j+1:]
		}
	}
	// snip footer
	if i := bytes.LastIndex(raw, snips[1]); i > -1 {
		if j := bytes.LastIndex(raw[:i], []byte{'\n'}); j > -1 {
			raw = raw[:j]
		}
	}
	return raw
}

func findTemplate() string {
	for _, dir := range build.Default.SrcDirs() {
		path := filepath.Join(dir, "github.com/mb0/lab/golab/gosrc")
		_, err := os.Stat(path)
		if err == nil {
			return path
		}
	}
	return ""
}
