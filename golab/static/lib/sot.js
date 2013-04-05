/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.

Package sot is a simple version of the operation transformation library:
ot.js (c) 2012-2013 Tim Baumann http://timbaumann.info MIT licensed.
*/
define(function() {

// Op represents a single operation.
// If op is number N it signifies:
// N > 0: Retain op bytes
// N < 0: Delete -op bytes
// B == 0: Noop
// If op is string S of utf8len N:
// N > 0: Insert string S
// N == 0: Noop

// Ops is a sequence of operations:
// [5, -2, "text"] // retain 5, delete 2, insert "text"

// javascript characters use UCS-2 encoding. we need utf-8 byte counts
function utf8len(str) {
	var i, c, n = 0;
	for (i=0; i<str.length; i++) {
		c = str.charCodeAt(i);
		if (c > 0x10000) n += 4;
		else if (c > 0x800) n += 3;
		else if (c > 0x80) n += 2;
		else n += 1;
	}
	return n;
}

// Count returns the number of retained, deleted and inserted bytes.
function count(ops) { // returns [ret, del, ins]
	var ret = 0, del = 0, ins = 0;
	for (var i=0; i < ops.length; i++) {
		var op = ops[i];
		if (typeof op == "string") {
			ins += utf8len(op);
		} else if (op < 0) {
			del += -op;
		} else if (op > 0) {
			ret += op;
		}
	}
	return [ret, del, ins];
}

// Merge attempts to merge consecutive operations the sequence.
function merge(ops) { // returns ops
	var lastop = 0;
	var res = [];
	for (var i=0; i < ops.length; i++) {
		var op = ops[i];
		if (!op) continue;
		var type = typeof op;
		if (type == typeof lastop && (
			type == "string" ||
			op > 0 && lastop > 0 ||
			op < 0 && lastop < 0)) {
			res[res.length-1] = lastop+op;
		} else {
			res.push(op);
		}
		lastop = res[res.length-1];
	}
	return res;
}

// Compose returns an operation sequence composed from the consecutive ops a and b.
// An error is returned if the composition failed.
function compose(a, b) { // returns [ab, err]
	if (!a || !b) {
		return [null, "Compose requires nonempty ops."];
	}
	var acount = count(a), bcount = count(b);
	if (acount[0]+acount[2] != bcount[0]+bcount[1]) {
		return [null, "Compose requires consecutive ops."];
	}
	var res = [], err = null;
	var ia = 0, ib = 0;
	var oa = a[ia++], ob = b[ib++];
	while (!!oa || !!ob) {
		var ta = typeof oa;
		if (ta == "number" && oa < 0) { // delete a
			res.push(oa);
			oa = a[ia++];
			continue;
		}
		var tb = typeof ob;
		if (tb == "string") { // insert b
			res.push(ob);
			ob = b[ib++];
			continue;
		}
		if (!oa || !ob || tb != "number") {
			return [null, "Compose encountered a short operation sequence."];
		}
		var od;
		if (ta == tb && oa > 0 && ob > 0) { // both retain
			od = oa - ob;
			if (od > 0) {
				oa -= ob;
				res.push(ob);
				ob = b[ib++];
			} else if (od < 0) {
				ob -= oa;
				res.push(oa);
				oa = a[ia++];
			} else {
				res.push(oa);
				oa = a[ia++];
				ob = b[ib++];
			}
		} else if (ta == "string" && ob < 0) { // insert delete
			od = utf8len(oa) + ob;
			if (od > 0) {
				oa = oa.substr(-ob);
				ob = b[ib++];
			} else if (od < 0) {
				ob = od;
				oa = a[ia++];
			} else {
				oa = a[ia++];
				ob = b[ib++];
			}
		} else if (ta == "string" && ob > 0) { // insert retain
			od = utf8len(oa) - ob;
			if (od > 0) {
				res.push(oa.substr(0, ob));
				oa = oa.substr(ob);
				ob = b[ib++];
			} else if (od < 0) {
				ob = -od;
				res.push(oa);
				oa = a[ia++];
			} else {
				res.push(oa);
				oa = a[ia++];
				ob = b[ib++];
			}
		} else if (ta == tb && oa > 0 && ob < 0) { // retain delete
			od = oa + ob;
			if (od > 0) {
				oa += ob;
				res.push(ob);
				ob = b[ib++];
			} else if (od < 0) {
				ob += oa;
				res.push(oa*-1);
				oa = a[ia++];
			} else {
				res.push(ob);
				oa = a[ia++];
				ob = b[ib++];
			}
		} else {
			alert("This should never have happened.");
		}
	}
	return [merge(res), err];
}

// Transform returns two operation sequences derived from the concurrent ops a and b.
// An error is returned if the transformation failed.
function transform(a, b) { // returns [a1, b1, err]
	if (!a || !b) {
		return [a, b, null];
	}
	var acount = count(a), bcount = count(b);
	if (acount[0]+acount[1] != bcount[0]+bcount[1]) {
		return [null, null, "Transform requires concurrent ops."];
	}
	var a1 = [], b1 = [], err = null;
	var ia = 0, ib = 0;
	var oa = a[ia++], ob = b[ib++];
	while (!!oa || !!ob) {
		var ta = typeof oa;
		if (ta == "string") { // insert a
			a1.push(oa);
			b1.push(utf8len(oa));
			oa = a[ia++];
			continue;
		}
		var tb = typeof ob;
		if (tb == "string") { // insert b
			a1.push(utf8len(ob));
			b1.push(ob);
			ob = b[ib++];
			continue;
		}
		if (!oa || !ob || ta != "number" || tb != ta) {
			return [null, null, "Compose encountered a short operation sequence."];
		}
		var od, om;
		if (oa > 0 && ob > 0) { // both retain
			od = oa - ob;
			if (od > 0) {
				om = ob;
				oa -= ob;
				ob = b[ib++];
			} else if (od < 0) {
				om = oa;
				ob -= oa;
				oa = a[ia++];
			} else {
				om = oa;
				oa = a[ia++];
				ob = b[ib++];
			}
			a1.push(om);
			b1.push(om);
		} else if (oa < 0 && ob < 0) { // both delete
			od = -oa + ob;
			if (od > 0) {
				oa -= ob;
				ob = b[ib++];
			} else if (od < 0) {
				ob -= oa;
				oa = a[ia++];
			} else {
				oa = a[ia++];
				ob = b[ib++];
			}
		} else if (oa < 0 && ob > 0) { // delete retain
			od = -oa - ob;
			if (od > 0) {
				om = -ob;
				oa += ob;
				ob = b[ib++];
			} else if (od < 0) {
				om = oa;
				ob += oa;
				oa = a[ia++];
			} else {
				om = oa;
				oa = a[ia++];
				ob = b[ib++];
			}
			a1.push(om);
		} else if (oa > 0 && ob < 0) { // retain delete
			od = oa + ob;
			if (od > 0) {
				om = ob;
				oa += ob;
				ob = b[ib++];
			} else if (od < 0) {
				om = -oa;
				ob += oa;
				oa = a[ia++];
			} else {
				om = -oa;
				oa = a[ia++];
				ob = b[ib++];
			}
			b1.push(om);
		} else {
			return [null, null, "Transform failed with incompatible operation sequences."];
		}
	}
	return [merge(a1), merge(b1), err];
}

return {
	utf8len: utf8len,
	count: count,
	merge: merge,
	compose: compose,
	transform: transform,
};
});
