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

return {
	crumbs: pathscrumbs,
};
});
