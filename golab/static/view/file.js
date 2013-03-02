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

return {View: FileView};
});
