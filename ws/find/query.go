// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package find

import (
	"fmt"
	"github.com/mb0/lab/ws"
	"regexp"
)

const (
	Exact = 0
	Start = 1 << iota
	End
	DblStart
	DblEnd
)

const (
	Resource = iota
	File
	Dir
)

type Query struct {
	Raw      string
	Terms    []Term
	Absolute bool
}

type Term struct {
	Words    []string
	Wildcard int
	Type     int
}

func (t Term) Equal(o Term) bool {
	if t.Type != o.Type || t.Wildcard != o.Wildcard {
		return false
	}
	if len(t.Words) != len(o.Words) {
		return false
	}
	for i, tw := range t.Words {
		if tw != o.Words[i] {
			return false
		}
	}
	return true
}

func (t Term) Compile() (*regexp.Regexp, error) {
	if len(t.Words) == 0 {
		return nil, nil
	}
	s := regexp.QuoteMeta(t.Words[0])
	for _, w := range t.Words[1:] {
		s = fmt.Sprintf("%s.*%s", s, regexp.QuoteMeta(w))
	}
	var start, end string
	if t.Wildcard&(Start|DblStart) != 0 {
		start = ".*"
	}
	if t.Wildcard&(End|DblEnd) != 0 {
		end = ".*"
	}
	return regexp.Compile(fmt.Sprintf("^%s%s%s$", start, s, end))
}

func Find(w *ws.Ws, query string) ([]*ws.Res, error) {
	q, err := new(parser).parse(query)
	if err != nil {
		return nil, err
	}
	var found, final []*ws.Res
	strict, nextStrict := q.Absolute, true
	// TODO if the query is absolute try looking up the starting resource
	iters := []*ws.Walker{ws.WalkAll(w)}
	terms := q.Terms
	for i := 0; i < len(terms); i++ {
		t := terms[i]
		re, err := t.Compile()
		if err != nil {
			return nil, err
		}
		for _, iter := range iters {
			for iter.Next() {
				r := iter.Res()
				if t.Type == Dir && r.Dir == nil || t.Type == File && r.Dir != nil {
					continue
				}
				match := false
				if re == nil {
					match = true
					if t.Wildcard == DblStart {
						nextStrict = false
						if i == len(terms)-1 {
							found = append(found, r)
							continue
						}
					}
				} else {
					match = re.MatchString(r.Name)
					switch t.Wildcard & (DblStart | DblEnd) {
					case DblStart:
						strict = false
						if match && i != len(terms)-1 {
							final = append(final, r)
						}
					case DblEnd:
						if match {
							final = append(final, r)
							nterm := Term{Wildcard: DblStart}
							if i == len(terms)-1 || !terms[i+1].Equal(nterm) {
								nterms := append(terms[:i], nterm)
								terms = append(nterms, terms[i:]...)
							}
						}
					}
				}
				if match {
					found = append(found, r)
				}
				if strict {
					iter.Skip()
				}
			}
		}
		if i == len(terms)-1 {
			break
		}
		iters = iters[:0]
		for _, dir := range found {
			if dir.Dir != nil {
				iters = append(iters, ws.Walk(dir))
			}
		}
		found = found[:0]
		strict, nextStrict = nextStrict, true
	}
	if len(final) > 0 {
		return append(final, found...), nil
	}
	return found, nil
}
