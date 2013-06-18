// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

angular.module("goapp.file", ["goapp.conn"])
.controller("FileCtrl", function($scope, $routeParams, $location) {
	var path = "/"+$routeParams.path, line = 0;
	if (path[path.length-1] == "/") {
		path = path.slice(0, path.length-1);
	}
	if ($location.hash().match(/^L\d+$/)) {
		line = parseInt($location.hash().slice(1), 10);
	}
	$scope.file = {path: path, line: line};
})
.directive("file", function() {
	return {
		restrict: "AE",
		replace: true,
		controller: "FileCtrl",
		template: [
			'<div class="file">path: {{ file.path }}, line {{ file.line }}</div>',
		].join(""),
	};
});