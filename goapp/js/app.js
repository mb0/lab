// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

require.config({
	paths: {
		angular: "//cdnjs.cloudflare.com/ajax/libs/angular.js/1.1.5/angular",
		ace: '/ace',
	},
	shim: {
		angular: {exports: "angular"},
	},
});

define(["modal", "angular", "conn", "tabs", "overview", "report", "file"], function(modal) {

angular.element(document).ready(function() {

angular.module("goapp", ["goapp.conn", "goapp.tabs", "goapp.overview", "goapp.report", "goapp.file"])
.config(function($routeProvider, $logProvider) {
	$logProvider.debugEnabled(false);
})
.run(function($rootScope, conn) {
	$rootScope.$on("conn.close", function(e, err) {
		var el = document.createElement("div");
		el.style.backgroundColor = "white";
		el.textContent = "connection closed";
		modal.show(el);
	});

	var proto = "ws:";
	if (location.protocol == "https:") {
		proto = "wss:";
	}
	conn.connect(proto +"//"+ location.host +"/ws");
});

angular.bootstrap(document, ['goapp']);
});

});
