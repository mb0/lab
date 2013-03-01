// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"bytes"
	"fmt"
	"github.com/mb0/lab/ws"
)

var (
	failmsg = "\x1b[41mFAIL\x1b[0m "
	failpre = "\x1b[41m    \x1b[0m "
	okmsg   = "\x1b[42mok  \x1b[0m "
	pending = "pending     "
)

type Report struct {
	Id   ws.Id
	Dir  string
	Path string
	Detail
}

func NewReport(pkg *Pkg) *Report {
	r := Report{pkg.Id, pkg.Dir, pkg.Path, pkg.Detail}
	r.Uses = make([]ws.Id, len(pkg.Uses))
	copy(r.Uses, pkg.Uses)
	r.Src.Info = pkg.Src.Info.Copy()
	r.Test.Info = pkg.Test.Info.Copy()
	return &r
}

func (r *Report) String() string {
	var buf bytes.Buffer
	for _, res := range []*Result{r.Src.Result, r.Test.Result} {
		if res == nil {
			continue
		}
		if buf.Len() > 0 {
			buf.WriteByte('\n')
		}
		if res.Err == nil {
			fmt.Fprintf(&buf, "%s%-7s %s", okmsg, res.Mode, r.Path)
			continue
		}
		fmt.Fprintf(&buf, "%s%-7s %s %s", failmsg, res.Mode, r.Path, res.Err)
		var b, l []byte
		for _, b = range [][]byte{[]byte(res.Stdout), []byte(res.Stderr)} {
			for b, l = line(b); len(l) > 0; b, l = line(b) {
				if len(l) == 1 || l[0] == '#' {
					continue
				}
				buf.WriteByte('\n')
				buf.WriteString(failpre)
				buf.Write(l)
			}
		}
	}
	if buf.Len() == 0 {
		return pending + r.Path
	}
	return buf.String()
}

func line(buf []byte) ([]byte, []byte) {
	if i := bytes.IndexByte(buf, '\n'); i > -1 {
		return buf[i+1:], buf[:i]
	}
	return nil, buf
}
