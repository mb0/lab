/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
angular.module("goapp.conn", [])
.service("conn",
function($rootScope, $location) {
	var c = this;
	c.url = "ws://"+ $location.host() +"/ws";
	c.url = "ws://localhost:8910/ws";
	c.conn = null;
	c.queue = [];
	c.debug = true;
	c.log = function(name, data) {
		if (c.debug) console.log(name, data);
	};
	c.trigger = function(name, data, head) {
		c.log(name, data);
		$rootScope.$emit(name, data);
	};
	c.connect = function() {
		c.log("conn.connect", c.url);
		c.conn = new WebSocket(c.url);
		c.conn.onopen = function(e) {
			c.trigger("conn.open", e);
			if (c.queue.length) {
				c.log("conn.work", c.queue);
				for (var i=0; i<c.queue.length; i++) {
					c.conn.send(c.queue[i]);
				}
				c.queue = [];
			}
		};
		c.conn.onclose = function(e) {
			c.conn = null;
			c.trigger("conn.close", e);
		};
		c.conn.onerror = function(e) {
			c.trigger("conn.error", e);
		};
		c.conn.onmessage = function(e) {
			var msg = JSON.parse(e.data);
			c.trigger("conn.msg", msg);
		};
	};
	c.connected = function() {
		return c.conn !== null;
	};
	c.send = function(head, data) {
		var msg = JSON.stringify({"Head": head, "Data": data});
		if (c.connected()) {
			c.log("conn.send", msg);
		} else {
			c.log("conn.queue", head);
			c.queue.push(msg);
		}
	};
});