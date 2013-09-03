// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

require.config({
	paths: {
		ace: '/static/ace',
	},
});

define(["ot"], function(ot) {

test("utf8len", function() {
	equal(ot.utf8len(""), 0, "empty length");
	equal(ot.utf8len("1"), 1, "simple length");
	equal(ot.utf8len("ä"), 2, "umlaut length");
});

test("ops count", function() {
	var o = [];
	function checklen(bl, tl) {
		var l, res = ot.count(o);
		ok(res.ret + res.del === bl, "base len");
		ok(res.ret + res.ins === tl, "target len");
	}
	checklen(0, 0);
	o.push(5);
	checklen(5, 5);
	o.push("abc");
	checklen(5, 8);
	o.push(2);
	checklen(7, 10);
	o.push(-2);
	checklen(9, 10);
});

test("ops merge", function() {
	var o = [5, 2, 0, "lo", "rem", 0, -3, -2, 0];
	deepEqual(ot.merge(o), [7, "lorem", -5]);
	deepEqual(ot.merge([1, 3, 1, -1]), [5, -1]);
});

var composeTests = [
	{
		a:  [3, -1],
		b:  [1, "tag", 2],
		ab: [1, "tag", 2, -1],
	},
	{
		a:  [1, "tag", 2],
		b:  [4, -2],
		ab: [1, "tag", -2],
	},
	{
		a:  [1, "tag"],
		b:  [2, -1, 1],
		ab: [1, "tg"],
	},
];
test("ops compose", function() {
	for (var i=0; i < composeTests.length; i++) {
		var c = composeTests[i];
		try {
			var res = ot.compose(c.a, c.b);
			deepEqual(res, c.ab);
		} catch (e) {
			equal(e, null, "error check");
		}
	}
});

var transformTests = [
	{
		a:  [1, "tag", 2],
		b:  [2, -1],
		a1: [1, "tag", 1],
		b1: [5, -1],
	},
	{
		a:  [1, "tag", 2],
		b:  [1, "tag", 2],
		a1: [1, "tag", 5],
		b1: [4, "tag", 2],
	},
	{
		a:  [1, -2],
		b:  [2, -1],
		a1: [1, -1],
		b1: [1],
	},
	{
		a:  [2, -1],
		b:  [1, -2],
		a1: [1],
		b1: [1, -1],
	},
];
test("ops transform", function() {
	for (var i=0; i < transformTests.length; i++) {
		var c = transformTests[i];
		var res = ot.transform(c.a, c.b);
		equal(res[2], null, "error check");
		deepEqual(res[0], c.a1);
		deepEqual(res[1], c.b1);
	}
});

});
