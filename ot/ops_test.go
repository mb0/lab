// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ot

import (
	"encoding/json"
	"testing"
)

func TestOpsCount(t *testing.T) {
	var o Ops
	checklen := func(bl, tl int) {
		ret, del, ins := o.Count()
		if l := ret + del; l != bl {
			t.Errorf("base len %d != %d", l, bl)
		}
		if l := ret + ins; l != tl {
			t.Errorf("taget len %d != %d", l, tl)
		}
	}
	checklen(0, 0)
	o = append(o, Op{N: 5})
	checklen(5, 5)
	o = append(o, Op{S: "abc"})
	checklen(5, 8)
	o = append(o, Op{N: 2})
	checklen(7, 10)
	o = append(o, Op{N: -2})
	checklen(9, 10)
}

func TestOpsMerge(t *testing.T) {
	o := Ops{
		{N: 5}, {N: 2}, {},
		{S: "lo"}, {S: "rem"}, {},
		{N: -3}, {N: -2}, {},
	}
	if mo := Merge(o); len(mo) != 3 {
		t.Errorf("got %+v", mo)
	}
}

func TestOpsEqual(t *testing.T) {
	var a, b Ops
	if !a.Equal(b) || !b.Equal(a) {
		t.Errorf("expect equal %v %v", a, b)
	}
	a = Ops{{N: 7}, {S: "lorem"}, {N: -5}}
	if a.Equal(b) || b.Equal(a) {
		t.Errorf("expect not equal %v %v", a, b)
	}
	b = Ops{{N: 7}, {S: "lorem"}, {N: -5}}
	if !a.Equal(b) || !b.Equal(a) {
		t.Errorf("expect equal %v %v", a, b)
	}
}

func TestOpsEncoding(t *testing.T) {
	e := `[7,"lorem",-5]`
	o := Ops{{N: 7}, {S: "lorem"}, {N: -5}}
	oe, err := json.Marshal(o)
	if err != nil {
		t.Error(err)
	}
	if string(oe) != e {
		t.Errorf("expected %s got %s", e, oe)
	}
	var eo Ops
	err = json.Unmarshal([]byte(e), &eo)
	if err != nil {
		t.Error(err)
	}
	if !o.Equal(eo) {
		t.Errorf("expected %v got %v", o, eo)
	}
}

var composeTests = []struct {
	a, b, ab Ops
}{
	{
		a:  Ops{{N: 3}},
		b:  Ops{{N: 1}, {S: "tag"}, {N: 2}},
		ab: Ops{{N: 1}, {S: "tag"}, {N: 2}},
	},
	{
		a:  Ops{{N: 1}, {S: "tag"}, {N: 2}},
		b:  Ops{{N: 4}, {N: -2}},
		ab: Ops{{N: 1}, {S: "tag"}, {N: -2}},
	},
}

func TestOpsCompose(t *testing.T) {
	for _, c := range composeTests {
		ab, err := Compose(c.a, c.b)
		if err != nil {
			t.Error(err)
		}
		if !ab.Equal(c.ab) {
			t.Errorf("expected %v got %v", c.ab, ab)
		}
	}
}

var transformTests = []struct {
	a, b, a1, b1 Ops
}{
	{
		a:  Ops{{N: 1}, {S: "tag"}, {N: 2}},
		b:  Ops{{N: 2}, {N: -1}},
		a1: Ops{{N: 1}, {S: "tag"}, {N: 1}},
		b1: Ops{{N: 5}, {N: -1}},
	},
	{
		a:  Ops{{N: 1}, {S: "tag"}, {N: 2}},
		b:  Ops{{N: 1}, {S: "tag"}, {N: 2}},
		a1: Ops{{N: 1}, {S: "tag"}, {N: 5}},
		b1: Ops{{N: 4}, {S: "tag"}, {N: 2}},
	},
}

func TestOpsTransform(t *testing.T) {
	for _, c := range transformTests {
		a1, b1, err := Transform(c.a, c.b)
		if err != nil {
			t.Error(err)
		}
		if !a1.Equal(c.a1) {
			t.Errorf("expected %v got %v", c.a1, a1)
		}
		if !b1.Equal(c.b1) {
			t.Errorf("expected %v got %v", c.b1, b1)
		}
	}
}
