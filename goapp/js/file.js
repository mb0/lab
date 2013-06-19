// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

angular.module("goapp.file", ["goapp.conn"])
.controller("FileCtrl", function($scope, $routeParams, $location, conn) {
	var path = "/"+$routeParams.path, line = 0;
	if (path[path.length-1] == "/") {
		path = path.slice(0, path.length-1);
	}
	if ($location.hash().match(/^L\d+$/)) {
		line = parseInt($location.hash().slice(1), 10);
	}
	$scope.childIcon = function(child) {
		if (!child.IsDir) {
			return 'icon-file-alt';
		}
		return 'icon-folder-close-alt';
	};
	$scope.fileHeader = function(file) {
		if (!file) return "";
		var path = file.Path;
		if (path[0] == "/") {
			path = path.substr(1);
		}
		var i, l = 0, parts = path.split("/");
		var result = ['<i class="'+(file.IsDir ? 'icon-folder-open-alt' : 'icon-file-alt')+'"></i> '];
		for (i=0; i < parts.length; i++) {
			l += parts[i].length;
			result.push('<a href="#/file/'+path.substr(0,l++)+'">/'+parts[i]+'</a>');
		}
		return result.join("");
	};
	var dereg = $scope.$on("conn.msg", function(e, msg) {
		if (msg.Head == "stat" && msg.Data.Path == path) {
			$scope.file = msg.Data;
			$scope.$digest();
		} else if (msg.Head == "stat.err" && msg.Data.Path == path) {
			$scope.file = msg.Data;
			$scope.$digest();
		}
	});
	function shorten(path) {
		var parts = path.split("/");
		if (parts.length > 1) {
			return parts[parts.length-2]+"/"+parts[parts.length-1];
		}
		return path;
	}
	$scope.tabs.add("/file/*path", {link: "/file"+path, name: shorten(path), close: true});
	conn.send("stat", path);
	$scope.$on("$destroy", dereg);
})
.directive("file", function() {
	return {
		restrict: "AE",
		replace: true,
		controller: "FileCtrl",
		template: [
			'<div class="file" ng-show="file">',
			'<div ng-bind-html-unsafe="fileHeader(file)"></div>',
			'<div ng-show="file.Error">error: path {{file.Path}} {{file.Error}}</div>',
			'<ul ng-show="file.IsDir"><li ng-repeat="child in file.Children">',
			'<a href="#/file{{ file.Path }}/{{ child.Name }}"><i class="{{childIcon(child)}}"></i> {{child.Name}}</a>',
			'</li></ul>',
			'<div ng-show="!file.IsDir">editor here</div>',
			'</div>',
		].join(""),
	};
});