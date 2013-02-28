/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(function() {
function newchild(pa, tag, inner) {
	var ele = document.createElement(tag);
	pa.appendChild(ele);
	if (inner) ele.innerHTML = inner;
	return ele;
}
function pad(str, l) {
	while (str.length < l) {
		str = "         " + str;
	}
	return str.slice(str.length-l);
}
function addreport(cont, r) {
	var c = newchild(cont, "div");
	header = '<span class="mode">'+r.Mode+'</span> '+ r.Path;
	if (r.Err != null) {
		c.setAttribute("class", "report fail");
		header = '<span class="status">FAIL</span> '+ header +": "+ r.Err;
	} else {
		c.setAttribute("class", "report ok");
		header = '<span class="status">ok</span> '+ header;
	}
	newchild(c, "header", header);
	var buf = [];
	if (r.Stdout) buf.push(r.Stdout);
	if (r.Stderr) buf.push(r.Stderr);
	var output = buf.join("\n")
	if (output) newchild(c, "pre", output);
	return c;
}
var cont = newchild(document.body, "div");
if (window["WebSocket"]) {
	var conn = new WebSocket("ws://"+ location.host+"/ws");
	conn.onclose = function(e) {
		newchild(cont, "div", "WebSocket closed.");
	};
	conn.onmessage = function(e) {
		var msg = JSON.parse(e.data);
		if (msg.Head == "report") {
			addreport(cont, msg.Data);
		} else {
			newchild(cont, "div", e.data);
		}
	};
	conn.onopen = function(e) {
		newchild(cont, "div", "WebSocket started.");
		conn.send('{"Head":"hi"}\n');
	};
} else {
	newchild(cont, "p", "WebSockets are not supported by your browser.");
}
});
