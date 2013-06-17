// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

angular.module("goapp.conn", [])
.factory("conn", function($rootScope, $log, $location) {
	function log(name, data) {
		$log.debug("conn."+name, data);
	}
	function trigger(name, data) {
		$log.debug("conn."+name, data);
		$rootScope.$emit("conn."+name, data);
	}
	var queue = [];
	var conn = {connected: false};
	conn.connect = function(url) {
		if (url === undefined) {
			url = "ws://"+ $location.host() +"/ws";
		}
		log("connect", url);
		ws = new WebSocket(url);
		ws.onopen = function(e) {
			conn.connected = true;
			trigger("open", e);
			if (queue.length) {
				log("work", queue);
				for (var i=0; i<queue.length; i++) {
					ws.send(JSON.stringify(queue[i]));
				}
				queue = [];
			}
		};
		ws.onclose = function(e) {
			ws = null;
			conn.connected = false;
			trigger("close", e);
		};
		ws.onerror = function(e) {
			trigger("error", e);
		};
		ws.onmessage = function(e) {
			trigger("msg", JSON.parse(e.data));
		};
	};
	conn.send = function(head, data) {
		var msg = {"Head": head, "Data": data};
		if (conn.connected) {
			log("send", msg);
			ws.send(JSON.stringify(msg));
		} else {
			log("queue", head);
			queue.push(msg);
		}
	};
	return conn;
});