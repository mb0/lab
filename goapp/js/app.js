// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

angular.module("goapp", ["goapp.conn", "goapp.report", "goapp.file", "goapp.tabs"])
.config(function($routeProvider, $logProvider) {
	$routeProvider.when("/about", {
		controller: "TabCtrl",
		template: [
			'<pre>',
			'<h3>golab</h3>'+
			'<a href="https://github.com/mb0/lab">github.com/mb0/lab</a> (c) Martin Schnabel '+
			'<a href="https://raw.github.com/mb0/lab/master/LICENSE">BSD License</a>',
			'</pre>'
		].join('\n'),
	})
	$logProvider.debugEnabled(true);
})
.run(function(conn) {
	conn.connect("ws://localhost:8910/ws");
});