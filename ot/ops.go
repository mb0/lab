// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ot provides operational transformation utilities for byte text collaboration.
//
// The code is based on, and in part compatible with
// https://github.com/Operational-Transformation/ot.js by Tim Baumann (MIT License).
package ot

import (
	"encoding/json"
	"fmt"
)

var noop Op

// Op represents a single operation.
type Op struct {
	// N signifies the operation type:
	// >  0: Retain  N bytes
	// <  0: Delete -N bytes
	// == 0: Noop or Insert string S
	N int
	S string
}

// MarshalJSON encodes op either as json number or string.
func (op *Op) MarshalJSON() ([]byte, error) {
	if op.N == 0 {
		return json.Marshal(op.S)
	}
	return json.Marshal(op.N)
}

// UnmarshalJSON decodes a json number or string into op.
func (op *Op) UnmarshalJSON(raw []byte) error {
	if len(raw) > 0 && raw[0] == '"' {
		return json.Unmarshal(raw, &op.S)
	}
	return json.Unmarshal(raw, &op.N)
}

// Ops represents a sequence of operations.
type Ops []Op

// Count returns the number of retained, deleted and inserted bytes.
func (ops Ops) Count() (ret, del, ins int) {
	for _, op := range ops {
		switch {
		case op.N > 0:
			ret += op.N
		case op.N < 0:
			del += -op.N
		case op.N == 0:
			ins += len(op.S)
		}
	}
	return
}

// Equal returns if other equals ops.
func (ops Ops) Equal(other Ops) bool {
	if len(ops) != len(other) {
		return false
	}
	for i, o := range other {
		if o != ops[i] {
			return false
		}
	}
	return true
}

// Merge attempts to merge consecutive operations in place and returns the sequence.
func Merge(ops Ops) Ops {
	o, l := -1, len(ops)
	for _, op := range ops {
		if op == noop {
			l--
			continue
		}
		var last Op
		if o > -1 {
			last = ops[o]
		}
		switch {
		case last.S != "" && op.N == 0:
			op.S = last.S + op.S
			l--
		case last.N < 0 && op.N < 0, last.N > 0 && op.N > 0:
			op.N += last.N
			l--
		default:
			o += 1
		}
		ops[o] = op
	}
	return ops[:l]
}

// getop returns the current sequence count and the next valid operation in ops or noop.
func getop(i int, ops Ops) (int, Op) {
	for ; i < len(ops); i++ {
		op := ops[i]
		if op != noop {
			return i + 1, op
		}
	}
	return i, noop
}

// sign return the sign of n.
func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}

// Compose returns an operation sequence composed from the consecutive ops a and b.
// An error is returned if the composition failed.
func Compose(a, b Ops) (ab Ops, err error) {
	reta, _, ins := a.Count()
	retb, del, _ := b.Count()
	if reta+ins != retb+del {
		err = fmt.Errorf("Compose requires consecutive ops.")
		return
	}
	if len(a) == 0 || len(b) == 0 {
		return
	}
	ia, oa := getop(0, a)
	ib, ob := getop(0, b)
	for oa != noop || ob != noop {
		if oa.N < 0 { // delete a
			ab = append(ab, oa)
			ia, oa = getop(ia, a)
			continue
		}
		if ob.N == 0 && ob.S != "" { // insert b
			ab = append(ab, ob)
			ib, ob = getop(ib, b)
			continue
		}
		if oa == noop || ob == noop {
			err = fmt.Errorf("Compose encountered a short operation sequence.")
			return
		}
		switch {
		case oa.N > 0 && ob.N > 0: // both retain
			switch sign(oa.N - ob.N) {
			case 1:
				oa.N -= ob.N
				ab = append(ab, ob)
				ib, ob = getop(ib, b)
			case -1:
				ob.N -= oa.N
				ab = append(ab, oa)
				ia, oa = getop(ia, a)
			default:
				ab = append(ab, oa)
				ia, oa = getop(ia, a)
				ib, ob = getop(ib, b)
			}
		case oa.N == 0 && ob.N < 0: // insert delete
			switch sign(len(oa.S) + ob.N) {
			case 1:
				oa = Op{S: string(oa.S[-ob.N:])}
				ib, ob = getop(ib, b)
			case -1:
				ob.N += len(oa.S)
				ia, oa = getop(ia, a)
			default:
				ia, oa = getop(ia, a)
				ib, ob = getop(ib, b)
			}
		case oa.N == 0 && ob.N > 0: // insert retain
			switch sign(len(oa.S) - ob.N) {
			case 1:
				ab = append(ab, Op{S: string(oa.S[:ob.N])})
				oa = Op{S: string(oa.S[ob.N:])}
				ib, ob = getop(ib, b)
			case -1:
				ob.N -= len(oa.S)
				ab = append(ab, oa)
				ia, oa = getop(ia, a)
			default:
				ab = append(ab, oa)
				ia, oa = getop(ia, a)
				ib, ob = getop(ib, b)
			}
		case oa.N > 0 && ob.N < 0: // retain delete
			switch sign(oa.N + ob.N) {
			case 1:
				oa.N += ob.N
				ab = append(ab, ob)
				ib, ob = getop(ib, b)
			case -1:
				ob.N += oa.N
				oa.N *= -1
				ab = append(ab, oa)
				ia, oa = getop(ia, a)
			default:
				ab = append(ab, ob)
				ia, oa = getop(ia, a)
				ib, ob = getop(ib, b)
			}
		default:
			panic("This should never have happened.")
		}
	}
	return
}

// Transform returns two operation sequences derived from the concurrent ops a and b.
// An error is returned if the transformation failed.
func Transform(a, b Ops) (a1, b1 Ops, err error) {
	if len(a) == 0 || len(b) == 0 {
		return
	}
	reta, dela, _ := a.Count()
	retb, delb, _ := b.Count()
	if reta+dela != retb+delb {
		err = fmt.Errorf("Transform requires concurrent ops.")
		return
	}
	ia, oa := getop(0, a)
	ib, ob := getop(0, b)
	for oa != noop || ob != noop {
		var om Op
		if oa.N == 0 && oa.S != "" { // insert a
			om.N = len(oa.S)
			a1 = append(a1, oa)
			b1 = append(b1, om)
			ia, oa = getop(ia, a)
			continue
		}
		if ob.N == 0 && ob.S != "" { // insert b
			om.N = len(ob.S)
			a1 = append(a1, om)
			b1 = append(b1, ob)
			ib, ob = getop(ib, b)
			continue
		}
		if oa == noop || ob == noop {
			err = fmt.Errorf("Transform encountered a short operation sequence.")
			return
		}
		switch {
		case oa.N > 0 && ob.N > 0: // both retain
			switch sign(oa.N - ob.N) {
			case 1:
				om.N = ob.N
				oa.N -= ob.N
				ib, ob = getop(ib, b)
			case -1:
				om.N = oa.N
				ob.N -= oa.N
				ia, oa = getop(ia, a)
			default:
				om.N = oa.N
				ia, oa = getop(ia, a)
				ib, ob = getop(ib, b)
			}
			a1 = append(a1, om)
			b1 = append(b1, om)
		case oa.N < 0 && ob.N < 0: // both delete
			switch sign(-oa.N + ob.N) {
			case 1:
				oa.N -= ob.N
				ib, ob = getop(ib, b)
			case -1:
				ob.N -= oa.N
				ia, oa = getop(ia, a)
			default:
				ia, oa = getop(ia, a)
				ib, ob = getop(ib, b)
			}
		case oa.N < 0 && ob.N > 0: // delete, retain
			switch sign(-oa.N - ob.N) {
			case 1:
				om.N = -ob.N
				oa.N += ob.N
				ib, ob = getop(ib, b)
			case -1:
				om.N = oa.N
				ob.N += oa.N
				ia, oa = getop(ia, a)
			default:
				om.N = oa.N
				ia, oa = getop(ia, a)
				ib, ob = getop(ib, b)
			}
			a1 = append(a1, om)
		case oa.N > 0 && ob.N < 0: // retain, delete
			switch sign(oa.N + ob.N) {
			case 1:
				om.N = ob.N
				oa.N += ob.N
				ib, ob = getop(ib, b)
			case -1:
				om.N = -oa.N
				ob.N += oa.N
				ia, oa = getop(ia, a)
			default:
				om.N = -oa.N
				ia, oa = getop(ia, a)
				ib, ob = getop(ib, b)
			}
			b1 = append(b1, om)
		default:
			err = fmt.Errorf("Transform failed with incompatible operation sequences.")
			return
		}
	}
	a1, b1 = Merge(a1), Merge(b1)
	return
}
