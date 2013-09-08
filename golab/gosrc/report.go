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
		if res.Errmsg == "" {
			fmt.Fprintf(&buf, "%s%-7s %s\n", okmsg, res.Mode, r.Path)
			continue
		}
		fmt.Fprintf(&buf, "%s%-7s %s %s\n", failmsg, res.Mode, r.Path, res.Errmsg)
		for _, l := range res.Output {
			if len(l) == 1 || l[0] == '#' {
				continue
			}
			buf.WriteString(failpre)
			buf.WriteString(l)
		}
	}
	if buf.Len() == 0 {
		return pending + r.Path
	}
	return buf.String()
}

type byDir []*Report

func (l byDir) Len() int {
	return len(l)
}
func (l byDir) Less(i, j int) bool {
	return l[i].Dir < l[j].Dir
}
func (l byDir) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
