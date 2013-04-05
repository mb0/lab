/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
require.config({
	paths: {
		json2: 'http://cdnjs.cloudflare.com/ajax/libs/json2/20121008/json2.min',
		zepto: 'http://cdnjs.cloudflare.com/ajax/libs/zepto/1.0rc1/zepto.min',
		underscore: 'http://cdnjs.cloudflare.com/ajax/libs/underscore.js/1.4.4/underscore-min',
		backbone: 'http://cdnjs.cloudflare.com/ajax/libs/backbone.js/0.9.10/backbone-min',
		ace: '/static/ace',
	},
	shim: {
		underscore: {exports: "_"},
		zepto: {exports: "$"},
		backbone: {exports: "Backbone", deps: ["underscore", "zepto"]},
	}
});

define(["lib/sot"], function(sot) {

test("ops count", function() {
	var o = [];
	function checklen(bl, tl) {
		var l, res = sot.count(o);
		ok(res[0] + res[1] === bl, "base len");
		ok(res[0] + res[2] === tl, "target len");
	}
	checklen(0, 0);
	o.push(5);
	checklen(5, 5);
	o.push("abc");
	checklen(5, 8);
	o.push(2);
	checklen(7, 10);
	o.push(-2)
	checklen(9, 10);
});

});
