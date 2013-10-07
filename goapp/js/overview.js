// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

define(["angular", "conn"], function() {

angular.module("goapp.overview", ["goapp.conn"])
.config(function($routeProvider) {
	$routeProvider.when("/overview", {
		controller: "TabCtrl",
		template: [
			'<div id="overview"><div overview></div>',
			'<pre class="about">',
			'<a href="https://github.com/mb0/lab">github.com/mb0/lab</a> (c) Martin Schnabel '+
			'<a href="https://raw.github.com/mb0/lab/master/LICENSE">BSD License</a>',
			'</pre></div>'
		].join('\n'),
	});
})
.controller("OverviewCtrl", function($scope, conn) {
	console.log("overview");
	var lastQuery = null;
	$scope.clear = function() {
		$scope.query = "";
		$scope.result = [];
		$scope.info = "";
	};
	$scope.clear();
	var dereg = $scope.$on("conn.msg", function(e, msg) {
		if (msg.Head == "find" && msg.Data.Query == lastQuery) {
			if (msg.Data.Error) {
				$scope.info = msg.Data.Error;
				$scope.result = [];
			} else {
				$scope.info = msg.Data.Query;
				$scope.result = msg.Data.Result;
			}
			$scope.$digest();
		}
	});
	$scope.find = function() {
		lastQuery = $scope.query;
		if (lastQuery) {
			$scope.info = "loading...";
			conn.send("find", lastQuery);
		} else {
			$scope.clear();
		}
	};
	$scope.$on("$destroy", dereg);
})
.directive("overview", function() {
	return {
		restrict: "AE",
		template: [
			'<div class="group"><h3>find</h3>',
			'<form ng-submit="find()"><input ng-model="query"><i class="icon-search" ng-click="find()"></i></form>',
			'<span>{{ info }}</span> <span>{{ result.length }} results</span>',
			'<i class="icon-remove" ng-click="clear()"></i>',
			'<ul><li ng-repeat="res in result">',
			'<a ng-href="#/file{{ res.Path }}" ng-bind-html-unsafe="res.Path"></a>',
			'</li></ul></div>',
			'<div class="group"><h3>tabs</h3>',
			'<ul><li ng-repeat="tab in tabs.list | filter:{close:true}">',
			'<a ng-href="#{{ tab.path }}" ng-bind-html-unsafe="tab.name"></a>',
			'<i class="icon-remove" ng-click="tabs.remove(tab)"></i>',
			'</li></ul></div>',
		].join(""),
		controller: "OverviewCtrl",
	};
});
});
