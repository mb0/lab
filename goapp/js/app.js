// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

require.config({
	paths: {
		angular: "//cdnjs.cloudflare.com/ajax/libs/angular.js/1.1.5/angular",
		ace: '/static/ace',
	},
	shim: {
		angular: {exports: "angular"},
	},
});

define(["angular", "conn", "tabs", "report", "file"], function(goapp) {

angular.module("goapp", ["goapp.conn", "goapp.tabs", "goapp.report", "goapp.file"])
.config(function($routeProvider, $logProvider) {
	$routeProvider.when("/about", {
		controller: "TabCtrl",
		template: [
			'<pre class="about">',
			'<h3>golab</h3>'+
			'<a href="https://github.com/mb0/lab">github.com/mb0/lab</a> (c) Martin Schnabel '+
			'<a href="https://raw.github.com/mb0/lab/master/LICENSE">BSD License</a>',
			'</pre>'
		].join('\n'),
	});
	$logProvider.debugEnabled(false);
})
.run(function(conn) {
	var proto = "ws:";
	if (location.protocol == "https:") {
		proto = "wss:";
	}
	conn.connect(proto +"//"+ location.host +"/ws");
});
});