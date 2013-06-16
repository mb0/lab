/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
angular.module("goapp", ["goapp.conn", "goapp.report"])
.config(
function($routeProvider) {
	$routeProvider
	.when("/about", {
		template: [
			'<pre>',
			'<h3>golab</h3>'+
			'<a href="https://github.com/mb0/lab">github.com/mb0/lab</a> (c) Martin Schnabel '+
			'<a href="https://raw.github.com/mb0/lab/master/LICENSE">BSD License</a>',
			'</pre>'
		].join('\n'),
	})
	.when("/report", {
		template: '<div id="report" report></div>',
	})
	.otherwise({
		redirectTo: "/",
	});
})
.run(function(conn, reportService) {
	conn.connect();
});