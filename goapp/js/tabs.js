/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
angular.module("goapp.tabs", [])
.run(function($rootScope, $route, $location) {
	var tabs = $rootScope.tabs = {list:[
		{path:"/about", name:'<i class="icon-beaker" title="about"></i>'},
		{path:"/report", name:'<i class="icon-circle" title="report"></i>'},
	], map:{}};
	for (var i=0; i < tabs.list.length; i++) {
		var tab = tabs.list[i];
		tabs.map[tab.path] = tab;
	}
	tabs.add = function(tab) {
		if (!tabs.map[tab.path]) {
			tabs.map[tab.path] = tab;
			tabs.list.push(tab);
		}
	};
	tabs.removeAt = function(idx) {
		var tab = tabs.list[idx];
		if (tab) {
			delete tabs.map[tab.path];
			tabs.list.splice(idx, 1);
		}
	};
	$rootScope.$on("$routeChangeSuccess", function(e, cur){
		if (tabs.activeTab) {
			tabs.activeTab.active = false;
		}
		tabs.activeTab = tabs.map[$location.path()];
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
			'<a href="#{{ tab.path }}" ng-bind-html-unsafe="tab.name"></a>',
			'<i class="icon-remove" ng-show="tab.close" ng-click="tabs.removeAt($index)"></i>',
			'</li></ul>',
		].join(""),
	};
});