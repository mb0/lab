// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"bytes"
	"fmt"
)

var (
	failmsg = "\x1b[41mFAIL\x1b[0m"
	failpre = "\x1b[41m    \x1b[0m"
	okmsg   = "\x1b[42mok  \x1b[0m"
)

type Report struct {
	Mode   string
	Path   string
	Start  int64
	Err    error
	Stdout string `json:",omitempty"`
	Stderr string `json:",omitempty"`
}

func (r *Report) String() string {
	if r.Err != nil {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%s %-7s %s %s", failmsg, r.Mode, r.Path, r.Err)
		var b, l []byte
		for _, b = range [][]byte{[]byte(r.Stdout), []byte(r.Stderr)} {
			for b, l = line(b); len(l) > 0; b, l = line(b) {
				if len(l) == 1 || l[0] == '#' {
					continue
				}
				fmt.Fprintf(&buf, "\n%s ", failpre)
				buf.Write(l)
			}
		}
		return buf.String()
	}
	return fmt.Sprintf("%s %-7s %s", okmsg, r.Mode, r.Path)
}

func line(buf []byte) ([]byte, []byte) {
	if i := bytes.IndexByte(buf, '\n'); i > -1 {
		return buf[i+1:], buf[:i]
	}
	return nil, buf
}
