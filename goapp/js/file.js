// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

define(["angular", "conn", "editor"], function(goapp) {

angular.module("goapp.file", ["goapp.conn", "goapp.editor"])
.config(function($routeProvider) {
	function shorten(path) {
		var parts = path.split("/");
		if (parts.length > 1) {
			return parts[parts.length-2]+"/"+parts[parts.length-1];
		}
		return path;
	}
	$routeProvider.when("/file/*path", {
		controller: "TabCtrl",
		factory: function(path) {
			return {path: path, name: shorten(path), close: true};
		},
		template: '<div file></div>',
	});
})
.controller("FileCtrl", function($scope, $routeParams, $location, conn) {
	var path = "/"+$routeParams.path, line = 0;
	if (path[path.length-1] == "/") {
		path = path.slice(0, path.length-1);
	}
	if ($location.hash().match(/^L\d+$/)) {
		line = parseInt($location.hash().slice(1), 10);
	}
	function fileHeader(file) {
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
	}
	$scope.openChild = function(c, e) {
		if (e && e.button !== 1) {
			var tab = $scope.tabs.get("/file"+path);
			if (tab) {
				tab.active = false;
				$scope.tabs.remove(tab);
			}
		}
		$location.path("/file"+ path +"/"+ c.Name);
	};
	var dereg = $scope.$on("conn.msg", function(e, msg) {
		if ((msg.Head == "stat" || msg.Head == "stat.err") && msg.Data.Path == path) {
			msg.Data.header = fileHeader(msg.Data);
			msg.Data.state = msg.Data.Error ? "error" : (msg.Data.IsDir ? "folder" : "file");
			$scope.file = msg.Data;
			$scope.$digest();
		}
	});
	conn.send("stat", path);
	$scope.$on("$destroy", dereg);
})
.filter("fileIcon", function() {
	return function(file) { return !file.IsDir ? "icon-file-alt" : "icon-folder-close-alt" };
})
.directive("file", function() {
	return {
		restrict: "AE",
		replace: true,
		controller: "FileCtrl",
		template: [
			'<div class="file" ng-show="file" ng-switch="file.state">',
			'<header ng-bind-html-unsafe="file.header"></header>',
			'<div ng-switch-when="error">error: path {{file.Path}} {{file.Error}}</div>',
			'<ul ng-switch-when="folder"><li ng-repeat="child in file.Children">',
			'<a href="" ng-click="openChild(child, $event)"><i class="{{ child|fileIcon }}"></i> {{child.Name}}</a>',
			'</li></ul>',
			'<div ng-switch-when="file" editor></div>',
			'</div>',
		].join(""),
	};
});
});