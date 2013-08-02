// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

define(["angular"], function(goapp) {

angular.module("goapp.tabs", [])
.controller("TabCtrl", function($scope, $route, $location) {
	$scope.tabs.activate($location.path(), $route.current);
})
.run(function($rootScope, $location) {
	var tabs = $rootScope.tabs = {};
	tabs.list = [
		{path:"/about", name:'<i class="icon-beaker" title="about"></i>'},
		{path:"/report", name:'<i class="icon-circle report" title="report"></i>'},
	];
	tabs.map = {};
	for (var i=0; i < tabs.list.length; i++) {
		var tab = tabs.list[i];
		tabs.map[tab.path] = tab;
	}
	tabs.history = ["/report"];
	tabs.removeHistory = function(path) {
		for (var i=0; i < tabs.history.length; i++) {
			if (tabs.history[i] == path) {
				tabs.history.splice(i, 1);
				break;
			}
		}
	};
	tabs.add = function(path, route) {
		var tab = route.factory(path);
		tabs.map[path] = tab;
		tabs.list.push(tab);
		return tab;
	};
	tabs.activate = function(path, route) {
		if (tabs.activeTab) {
			tabs.activeTab.active = false;
			tabs.activeTab = null;
		}
		var tab = tabs.map[path];
		if (!tab) {
			tab = tabs.add(path, route);
		} else {
			tabs.removeHistory(path);
		}
		tabs.history.push(path);
		tab.active = true;
		tabs.activeTab = tab;
	};
	tabs.get = function(path) {
		for (var i=0; i < tabs.list.length; i++) {
			var tab = tabs.list[i];
			if (tab.path == path) {
				return tab;
			}
		}
		return null;
	};
	tabs.remove = function(tab) {
		if (!tab) return;
		var idx = -1;
		for (i=0; i < tabs.list.length; i++) {
			if (tabs.list[i] == tab) {
				idx = i;
				break;
			}
		}
		if (idx < 0) return;
		delete tabs.map[tab.path];
		tabs.list.splice(idx, 1);
		tabs.removeHistory(tab.path);
		if (tab.active) {
			tabs.activeTab = null;
			if (tabs.history.length > 0) {
				$location.path(tabs.history[tabs.history.length-1]);
			}
		}
	};
})
.directive("tabs", function() {
	return {
		restrict: "EA",
		template: [
			'<ul><li ng-repeat="tab in tabs.list" ng-class="{active: tab.active, error: tab.error}" ng-switch="tab.close">',
			'<a ng-href="#{{ tab.path }}" ng-bind-html-unsafe="tab.name"></a>',
			'<i class="icon-remove" ng-switch-when="true" ng-click="tabs.remove(tab)"></i>',
			'</li></ul>',
		].join(""),
	};
});
});