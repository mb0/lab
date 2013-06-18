// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

angular.module("goapp.report", ["goapp.conn"])
.run(function($rootScope) {
	var r = $rootScope.reports = {map:{}, list:[]};
	function fix(report) {
		report.Res = report.Test.Result || report.Src.Result;
		report.Status = report.Res.Err ? "fail" : "ok";
		var out = (report.Res.Stdout || "") + (report.Res.Stderr || "");
		out = out.replace(/(\/([^\/\s]+\/)+(\S+?\.go))\:(\d+)(?:\:(\d+))?\:/g, '<a href="#/file$1#L$4">$2$3:$4</a>');
		out = out.replace(/\n(([\w_]+\.go)\:(\d+)(?:\:\d+)?\:)/g, '\n<a href="#/file' + report.Dir + '/$2#L$3">$1</a>');
		report.Output = out.replace(/(^(#.*|\S)\n|\n#[^\n]*)/g, "");
	}
	function add(reports) {
		var i, id, old, report;
		for (i=0; i < reports.length; i++) {
			report = reports[i];
			fix(report);
			old = r.map[report.Id];
			report.Detail = old ? old.Detail : false;
			r.map[report.Id] = report;
		}
		r.list.length = 0;
		for (id in r.map) {
			r.list.push(r.map[id]);
		}
		$rootScope.$digest();
	}
	$rootScope.$on("conn.msg", function(e, msg) {
		if (msg.Head == "reports") add(msg.Data);
		else if (msg.Head == "report") add([msg.Data]);
	});
})
.filter("reportIcon", function() {
	return function(detail) { return detail ? "icon-minus" : "icon-plus" };
})
.directive("report", function() {
	return {
		restrict: "AE",
		replace: true,
		template: [
			'<ul><li class="report report-{{r.Status}}" ng-repeat="r in reports.list | orderBy:\'Path\'">',
			'<span class="status">{{r.Status|uppercase}} <i ng-show="r.Output" ng-click="r.Detail = !r.Detail" class="{{ r.Detail|reportIcon }}"></i></span>',
			'<span>{{r.Res.Mode}}</span> {{r.Path}}',
			'<pre ng-show="r.Output && r.Detail" ng-bind-html-unsafe="r.Output"></pre>',
			'</li></ul>',
		].join(""),
	};
});