// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

define(["angular", "otdoc", "acecfg", "conn"], function(angular, otdoc, acecfg) {
	
angular.module("goapp.editor", ["goapp.conn"])
.run(function($rootScope, conn) {
	var docs = $rootScope.docs = {
		map: {}, // by id
	};
	docs.subscribe = function(id, path, handler) {
		var doc = docs.map[id];
		if (doc !== undefined) {
			console.log("known document", path);
			handler(doc);
		}
		doc = new otdoc.Doc(id, path);
		docs.map[id] = doc;
		var listener = function() {
			doc.Ace.removeListener("init", listener);
			handler(doc);
		};
		doc.Ace.on("init", listener);
		conn.send("subscribe", {Id: id});
	};
	$rootScope.$on("conn.msg", function(e, msg) {
		var doc;
		if (msg.Head == "subscribe") {
			doc = docs.map[msg.Data.Id];
			if (!doc) {
				console.log("subscribe unknown document", msg.Data);
				return;
			}
			var text = msg.Data.Ops && msg.Data.Ops[0] || "";
			doc.init(msg.Data.Rev, msg.Data.User, text);
			doc.Ace.on("ops", function(e) {
				conn.send("revise", {Id: doc.Id, Rev: doc.Rev, Ops: e.ops});
			});
		} else if (msg.Head == "revise") {
			doc = docs.map[msg.Data.Id];
			if (!doc) {
				console.log("revise unknown document", msg.Data);
				return;
			}
			try {
				if (doc.User === msg.Data.User) {
					doc.ackOps(msg.Data.Ops);
				} else {
					doc.recvOps(msg.Data.Ops);
				}
			} catch (err) {
				console.log(err, msg.Data);
				alert("revise panic "+err);
			}
		} else if (msg.Head == "revise.err") {
			console.log(msg.Data);
			alert("doc panic "+ msg.Data.Err);
		} else if (msg.Head == "publish") {
			doc = docs.map[msg.Data.Id];
			if (!doc) {
				console.log("publish unknown document", msg.Data);
				return;
			}
			doc.Status = "published";
		} else if (msg.Head == "unsubscribe") {
			doc = docs.map[msg.Data.Id];
			if (!doc) {
				console.log("unsubscribe unknown document", msg.Data);
				return;
			}
			delete docs.map[msg.Data.Id];
		}
	});
})
.controller("EditorCtrl", function($scope, $element, conn) {
	// TODO subscribe to document and listen for changes
	var editor = null;
	var el = document.createElement("div");
	$element.append(el);
	$scope.docs.subscribe($scope.file.Id, $scope.file.Path, function(doc){
		$scope.doc = doc;
		var renderer = acecfg.createRenderer(el);
		var session = acecfg.createSession(doc.Ace, "ace/mode/text");
		editor = acecfg.createEditor(renderer, session, true);
	});
	$scope.$on("conn.msg", function(e, msg) {
		if (msg.Head == "complete" && msg.Data.Id === $scope.doc.Id) {
			// TODO show completion popup
			console.log("complete");
		}
	});
})
.directive("editor", function() {
	return {
		restrict: "EA",
		controller: "EditorCtrl",
	};
});
});