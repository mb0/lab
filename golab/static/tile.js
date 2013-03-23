/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["backbone"], function() {

var Tile = Backbone.Model.extend({
	defaults: {
		id: null,
		uri: null,
		name: "unnamed",
		close: false,
		active: false,
		view: null
	}
});

var Tiles = Backbone.Collection.extend({
	model: Tile
});

return {
	Tile: Tile,
	Tiles: Tiles,
};
});
