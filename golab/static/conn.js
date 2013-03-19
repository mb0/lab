/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["json2", "backbone"], function() {

if (!window["WebSocket"]) {
	return null;
}

var Conn = Backbone.Model.extend({
	constructor: function(data, opts) {
		this.wsconn = null;
		this.queue = [];
		this.debug = false;
		Backbone.Model.prototype.constructor.call(this, data, opts);
	},
	_trigger: function(name, data) {
		if (this.debug) console.log(name, data);
		this.trigger(name, data);
	},
	connect: function() {
		var c = this;
		if (c.urlRoot) {
			c.wsurl = c.url();
		}
		if (!c.wsurl) {
			c.wsurl = "ws://"+ location.host +"/ws";
		}
		c._trigger("connect", c.wsurl);
		var ws = c.wsconn = new WebSocket(c.wsurl);
		ws.onopen = function(e) {
			c._trigger("open", e);
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
			c._trigger("close", e);
		};
		ws.onmessage = function(e) {
			var msg = JSON.parse(e.data)
			c._trigger("msg", msg);
			c.trigger("msg:"+ msg.Head, msg.Data);
		};
		ws.onerror = function(e) {
			c._trigger("error", e.message);
		};
	},
	connected: function() {
		return this.wsconn != null;
	},
	send: function(head, data) {
		var msg = JSON.stringify({"Head":head, "Data":data});
		if (this.wsconn != null) {
			if (this.debug) console.log("send", msg);
			this.wsconn.send(msg);
		} else {
			if (this.debug) console.log("push", head);
			this.queue.push(msg);
		}
	},

});

return new Conn();

});
