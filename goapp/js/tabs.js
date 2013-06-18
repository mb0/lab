/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
angular.module("goapp.tabs", [])
.run(function($rootScope) {
	$rootScope.tabs = [
		{link:"#/about", name:"about"},
		{link:"#/report", name:"report"},
		{link:"#/file/a/b", name:"a/b", close: true},
	];
})
.directive("tabs", function() {
	return {
		restrict: "EA",
		template: [
			'<ul><li ng-repeat="tab in tabs" ng-class="{active: tab.active}">',
			'<a href="{{ tab.link }}">{{ tab.name }}</a>',
			'</li></ul>',
		].join(""),
	};
});