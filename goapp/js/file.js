// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

angular.module("goapp.file", [])
.controller("FileCtrl", function($scope, $log, $routeParams) {
	$log.info("filepath:", "/"+$routeParams.path);
});