/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define([], function() {

function pathscrumbs(path) {
	if (!path) return [];
	var i = 0;
	if (path[0] == "/") path = path.substr(1);
	return _.map(path.split("/"), function(p){
		i += p.length;
		return [path.substr(0, i++), p];
	});
}

function shorten(path, num) {
	var crumbs = pathscrumbs(path);
	return _.map(_.last(crumbs, num), function(p) {
		return p[1];
	}).join("/") || path;
}

function splitHashLine(path) {
	var line = 0, pathline = path.split("#L");
	if (pathline.length > 1 && pathline[1].match(/^\d+$/)) {
		path = pathline[0], line = parseInt(pathline[1], 10);
	}
	if (path && path[path.length-1] == "/") {
		path = path.slice(0, path.length-1);
	}
	return {path: path, line: line};
}

return {
	crumbs: pathscrumbs,
	shorten: shorten,
	splitHashLine: splitHashLine,
};
});
