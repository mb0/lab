// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package find

import (
	"testing"
)

var parserTests = []struct {
	raw   string
	query Query
}{
	// resources named "a"
	{
		`a`,
		Query{Terms: []Term{{Words: []string{`a`}}}},
	},
	// resources starting with "a" containing "b" ending with "c"
	{
		`a*b*c`,
		Query{Terms: []Term{{Words: []string{`a`, `b`, `c`}}}},
	},
	// resources with prefix "x."
	{
		`a*`,
		Query{Terms: []Term{{Words: []string{`a`}, Wildcard: End}}},
	},
	// resources with suffix ".go"
	{
		`*a`,
		Query{Terms: []Term{{Words: []string{`a`}, Wildcard: Start}}},
	},
	// same as query "a" but does not match resources at depth 1
	{
		`/**/a`,
		Query{Absolute: true, Terms: []Term{{Wildcard: DblStart, Type: Dir}, {Words: []string{`a`}}}},
	},
	// resources containing "a"
	{
		`*a*`,
		Query{Terms: []Term{{Words: []string{`a`}, Wildcard: Start | End}}},
	},
	// resources named "special\*$chars"
	{
		`special\\\*\$chars`,
		Query{Terms: []Term{{Words: []string{`special\*$chars`}}}},
	},
	// resources named "a"
	{
		`**/a`,
		Query{Terms: []Term{{Wildcard: DblStart, Type: Dir}, {Words: []string{`a`}}}},
	},
	// files named "a"
	{
		`a$`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: File}}},
	},
	// dirs named "a"
	{
		`a/`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: Dir}}},
	},
	// children of "dir"
	{
		`a/*`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: Dir}, {Wildcard: Start}}},
	},
	// grand children of "dir"
	{
		`a/*/*`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: Dir}, {Wildcard: Start, Type: Dir}, {Wildcard: Start}}},
	},
	// descendant of "dir"
	{
		`a/**`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: Dir}, {Wildcard: DblStart}}},
	},
	// a/*b || a/**/*b
	{
		`a/**b`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: Dir}, {Words: []string{`b`}, Wildcard: DblStart}}},
	},
	// a/b* || a/b*/**
	{
		`a/b**`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: Dir}, {Words: []string{`b`}, Wildcard: DblEnd}}},
	},
	// a/*b* || a/**/*b*
	{
		`a/**b*`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: Dir}, {Words: []string{`b`}, Wildcard: DblStart | End}}},
	},
	// a/*b* || a/*b*/**
	{
		`a/*b**`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: Dir}, {Words: []string{`b`}, Wildcard: Start | DblEnd}}},
	},
	// *a*/b || *a*/**/b
	{
		`*a**/b`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: Dir, Wildcard: Start | DblEnd}, {Words: []string{`b`}}}},
	},
	// *a*/b
	{
		`**a*/b`,
		Query{Terms: []Term{{Words: []string{`a`}, Type: Dir, Wildcard: DblStart | End}, {Words: []string{`b`}}}},
	},
}

func TestParseQuery(t *testing.T) {
	var p parser
	for _, test := range parserTests {
		q, err := p.parse(test.raw)
		if err != nil {
			t.Errorf("query %q err %s", test.raw, err)
		}
		if q.Absolute != test.query.Absolute {
			t.Errorf("query %q is absolute %+v is not", test.raw, q)
		}
		if len(q.Terms) != len(test.query.Terms) {
			t.Errorf("query %q terms differ %+v", test.raw, q)
			continue
		}
		for i, term := range test.query.Terms {
			if !q.Terms[i].Equal(term) {
				t.Errorf("query %q at %d expect %+v got %+v", test.raw, i, term, q.Terms[i])
			}
		}
	}
}
