/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
angular.module("goapp.tabs", [])
.run(function($rootScope, $route, $location) {
	var tabs = $rootScope.tabs = {list:[]};
	for (var path in $route.routes) {
		var r = $route.routes[path];
		if (!r.tabs) continue;
		for (var link in r.tabs) {
			var tab = r.tabs[link];
			tab.link = link;
			tabs.list.push(tab);
		}
	}
	$rootScope.$on("$routeChangeSuccess", function(e, cur){
		if (tabs.activeTab) {
			tabs.activeTab.active = false;
		}
		tabs.activeTab = cur.$$route.tabs[$location.path()];
		if (tabs.activeTab) {
			tabs.activeTab.active = true;
		}
	});
})
.directive("tabs", function() {
	return {
		restrict: "EA",
		template: [
			'<ul><li ng-repeat="tab in tabs.list" ng-class="{active: tab.active}">',
			'<a href="#{{ tab.link }}" ng-bind-html-unsafe="tab.name"></a>',
			'</li></ul>',
		].join(""),
	};
});