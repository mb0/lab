/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["base", "lib/paths", "backbone"],
function(base, paths) {

var File = Backbone.Model.extend({
	idAttribute: "Id",
	getPath: function() {
		var path = this.get("Parent").get("Path");
		if (path && path[path.length-1] != "/") {
			path += "/";
		}
		return path + this.get("Name");
	},
	getIcon: function(open) {
		if (!this.get("IsDir")) {
			return 'icon-file';
		}
		return 'icon-folder-close-alt';
	},
});

var Files = Backbone.Collection.extend({model:File});

var FileListItem = base.ListItemView.extend({
	template: _.template([
		'<a href="#file<%- getPath() %>">',
		'<i class="<%- getIcon() %>"></i> <%- get("Name") %>',
		'</a>'
	].join('')),
});

var FileList = base.ListView.extend({
	itemView: FileListItem,
});

var Folder = File.extend({
	getPath: function() {
		return this.get("Path");
	},
	getIcon: function() {
		return 'icon-folder-open-alt';
	},
	getCrumbs: function() {
		return _.map(paths.crumbs(this.getPath()), function(c) {
			return '<a href="#file/'+ c[0] +'">/'+c[1]+"</a>";
		}).join("");
	},
});

var FolderView = Backbone.View.extend({
	template: _.template(
		'<header><i class="<%- getIcon() %>"></i> <%= getCrumbs() %></header>'
	),
	events: {
		"click a": "navigate",
	},
	initialize: function(opts) {
		this.tile = opts.tile;
		this.$el.addClass("folder");
		this.model = this.model.isValid ? this.model : new Folder(this.model);
		this.children = new Files();
		this.list = new FileList({collection:this.children});
		this.$el.append($("<header>")).append(this.list.el);
		this.listenTo(this.model, "change", this.render);
		this.listenTo(this.model, "remove", this.remove);
		this.listenTo(this.tile, "remove", this.remove);
		this.render();
	},
	render: function() {
		this.$("header").replaceWith(this.template(this.model));
		this.children.reset(_.map(this.model.get("Children"), function(c) {
			c.Parent = this.model;
			return c;
		}, this));
		return this;
	},
	navigate: function(e) {
		var href = $(e.currentTarget).attr("href");
		if (!href || href.indexOf("http") === 0) return;
		e.preventDefault();
		if (href == "#"+ this.tile.get("uri")) return;
		if (e.button !== 1) { // middle click
			this.tile.collection.remove(this.tile);
		}
		Backbone.history.navigate(href, {trigger: true});
	},
});

return {
	View: FolderView,
};
});
