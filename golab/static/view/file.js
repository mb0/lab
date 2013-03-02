/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["base", "conn"], function(base, conn) {

var File = Backbone.Model.extend({
	idAttribute: "Id",
});

var Files = Backbone.Collection.extend({model:File});

var FileListItem = base.ListItemView.extend({
	template: _.template('<a><%- get("Id") %></a>'),
});

var FileList = base.ListView.extend({
	itemView: FileListItem,
});

var FileView = Backbone.View.extend({
	tagName: "section",
	initialize: function() {
		this.files = new Files([{Id:1}, {Id:2}, {Id:3}]);
		this.listview = new FileList({collection:this.files});
		this.render();
	},
	render: function() {
		this.$el.append(this.listview.render().$el);
		return this;
	},
});

var views = {};
function openfile(path) {
	if (!path) {
		// show src dirs
		return;
	}
	if (path[path.length-1] == "/") {
		path = path.slice(0, path.length-1);
	}
	path = "/"+path;
	var view = views[path];
	if (!view) {
		view = new FileView({id: _.uniqueId("file")});
		views[path] = view;
	}
	return {
		id: view.id,
		uri: "file"+path,
		name: path,
		view: view,
		active: true,
		closable: true,
	};
};

return {
	View: FileView,
	router: {
		route:    "file/*path",
		name:     "openfile",
		callback: openfile,
	},
};
});
