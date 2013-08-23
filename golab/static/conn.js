/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["json2", "underscore", "backbone"], function() {

if (!window.WebSocket) {
	return null;
}
var Conn = function(){
	this.wsconn = null;
	this.wsurl = null;
	this.queue = [];
	this.debug = false;
	_.extend(this, Backbone.Events);
};
Conn.prototype = {
	log: function(name, data) {
		if (this.debug) {
			console.log(name, data);
		}
	},
	connect: function() {
		var c = this;
		if (!c.wsurl) {
			var proto = "ws:";
			if (location.protocol == "https:") {
				proto = "wss:";
			}
			c.wsurl = proto +"//"+ location.host +"/ws";
		}
		c.log("connect", c.wsurl);
		c.trigger("connect", c.wsurl);
		var ws = c.wsconn = new WebSocket(c.wsurl);
		ws.onopen = function(e) {
			c.log("open", e);
			c.trigger("open", e);
			if (c.queue) {
				if (c.debug) console.log("work", c.queue);
				_.each(c.queue, function(e) {
					c.wsconn.send(e);
				});
				c.queue = [];
			}
		};
		ws.onclose = function(e) {
			c.wsconn = null;
			c.log("close", e);
			c.trigger("close", e);
		};
		ws.onmessage = function(e) {
			var msg = JSON.parse(e.data);
			c.log("msg", msg);
			c.trigger("msg", msg);
			c.trigger("msg:"+ msg.Head, msg.Data);
		};
		ws.onerror = function(e) {
			c.log("error", e.message);
			c.trigger("error", e.message);
		};
	},
	connected: function() {
		return this.wsconn !== null;
	},
	send: function(head, data) {
		var msg = JSON.stringify({"Head":head, "Data":data});
		if (this.wsconn !== null) {
			this.log("send", msg);
			this.wsconn.send(msg);
		} else {
			this.log("push", head);
			this.queue.push(msg);
		}
	},

};

return new Conn();

});
