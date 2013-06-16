/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
angular.module("goapp.report", [])
.service("reportService",
function($rootScope) {
	var s = this;
	var cache = {}, list = [], scopes = {};
	function add(report) {
		report.Res = report.Test.Result || report.Src.Result;
		report.Status = report.Res.Err ? "fail" : "ok";
		report.Output = (report.Res.Stdout || "") + (report.Res.Stderr || "");
		var i = cache[report.Id];
		if (i !== undefined) {
			report.Detail = list[i].Detail;
		} else {
			i = cache[report.Id] = list.length;
			report.Detail = false;
		}
		list[i] = report;
	}
	$rootScope.$on("conn.msg", function(e, msg) {
		if (msg.Head === "reports")
			for (var i=0; i < msg.Data.length; i++)
				add(msg.Data[i]);
		else if (msg.Head === "report")
			add(msg.Data);
		else return;
		for (var id in scopes)
			scopes[id].$digest();
	});
	this.register = function(scope) {
		scopes[scope.$id] = scope;
		scope.$on("$destroy", function() {
			delete scopes[scope.$id];
		});
		return list;
	};
})
.controller("ReportCtrl",
function($scope, $element, reportService) {
	$scope.reports = reportService.register($scope);
	$scope.predicate = ["Path"];
	$scope.getIcon = function(report) {
		return report.Detail ? "icon-minus" : "icon-plus";
	};
})
.directive("report",
function() {
	return {
		restrict: "AE",
		replace: true,
		controller: "ReportCtrl",
		template: [
			'<ul><li class="report report-{{r.Status}}" ng-repeat="r in reports | orderBy:predicate">',
			'<span class="status">{{r.Status|uppercase}} <i ng-show="r.Output" ng-click="r.Detail = !r.Detail" class="{{ getIcon(r) }}"></i></span>',
			'<span>{{r.Res.Mode}}</span> {{r.Path}}',
			'<pre ng-show="r.Output && r.Detail">{{r.Output}}</pre>',
			'</li></ul>',
		].join(""),
	};
});