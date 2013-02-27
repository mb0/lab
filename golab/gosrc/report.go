// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosrc

import (
	"bytes"
	"fmt"
	"time"
)

var (
	failmsg = "\x1b[41mFAIL\x1b[0m"
	failpre = "\x1b[41m    \x1b[0m"
	okmsg   = "\x1b[42mok  \x1b[0m"
)

type Report struct {
	Mode, Path string
	Start      time.Time
	Err        error
	out, err   bytes.Buffer
}

func (r *Report) String() string {
	if r.Err != nil {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%s %-7s %s %s", failmsg, r.Mode, r.Path, r.Err)
		for _, b := range []*bytes.Buffer{&r.out, &r.err} {
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
	return fmt.Sprintf("%s %-7s %s", okmsg, r.Mode, r.Path)
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
