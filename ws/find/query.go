// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package find

import (
	"github.com/mb0/lab/ws"
	"strings"
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
	Word     string
	Wildcard int
	Type     int
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
		for _, iter := range iters {
			for iter.Next() {
				r := iter.Res()
				switch t.Type {
				case Dir:
					if r.Dir == nil {
						continue
					}
				case File:
					if r.Dir != nil {
						continue
					}
				}
				match := false
				if t.Word == "" {
					match = true
					if t.Wildcard == DblStart {
						nextStrict = false
						if i == len(terms)-1 {
							found = append(found, r)
							continue
						}
					}
				} else {
					switch t.Wildcard {
					case Exact:
						match = r.Name == t.Word
					case Start:
						match = strings.HasSuffix(r.Name, t.Word)
					case DblStart:
						strict = false
						match = strings.HasSuffix(r.Name, t.Word)
						if match && i != len(terms)-1 {
							final = append(final, r)
						}
					case End, DblEnd:
						match = strings.HasPrefix(r.Name, t.Word)
						if match && t.Wildcard == DblEnd {
							strict = false
							nextStrict = false
						}
					case Start | End, DblStart | End, Start | DblEnd:
						match = strings.Contains(r.Name, t.Word)
						if !match {
							break
						}
						if t.Wildcard == DblStart|End {
							strict = false
						}
						if t.Wildcard == Start|DblEnd {
							nextStrict = false
							final = append(final, r)
							nterm := Term{Wildcard: DblStart}
							nterms := terms[:i]
							if i == len(terms)-1 || terms[i+1] != nterm {
								nterms = append(nterms, nterm)
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
