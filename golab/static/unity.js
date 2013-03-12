/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(function() {
try {
	var Unity = external.getUnityObject(1.0);
	var baseurl = [location.protocol, "//", location.host, "/"].join("");
	Unity.init({
		name: "Golab",
		homepage: baseurl,
		iconUrl: (baseurl + "static/golab.png"),
		onInit: function() {
			console.log("unity ready");
		}
	});
} catch (e) {
	console.log("failed to initialize unity web app api"+ e);
}
});
