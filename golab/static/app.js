/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["base"], function(base) {

var Tile = Backbone.Model.extend({
	defaults: {
		id: null,
		uri: null,
		name: "unnamed",
		closable: false,
		active: false,
		view: null
	}
});

var Tiles = Backbone.Collection.extend({
	model: Tile
});

var Tabs = base.ListView.extend({
	itemView: base.ListItemView.extend({
		template: _.template([
			'<a href="#<%- get("uri") %>" <% if (get("active")) { %> class="active" <% } %>>',
			'<%= get("name") %></a>',
			'<% if (get("closable")) { %>',
			'<sup class="close">x</sup>',
			'<% } %>',
		].join('')),
		events: { "click .close": "removeModel"},
		removeModel: function() {
			this.collection.remove(this.model);
		}
	})
});

var App = Backbone.View.extend({
	el: $("#app").get(0),
	initialize: function() {
		this.active = null;
		this.tabs = new Tabs({collection: this.collection}).render();
		$('<nav>').appendTo(this.$el).append(this.tabs.$el);
		this.$cont = $('<div>').appendTo(this.$el);
		this.listenTo(this.collection, "reset remove", this.render);
	},
	activate: function(id) {
		if (this.active != null) {
			if (this.active.get("id") == id) return;
			this.active.set("active", false);
			this.$cont.children().remove();
		}
		this.active = this.collection.get(id);
		if (this.active != null) {
			this.active.set("active", true);
			this.$cont.append(this.active.get("view").$el);
		}
	},
	render: function() {
		var active = this.collection.find(function(model) {
			return model.active;
		});
		this.activate(active && active.get("id") || null);
		return this;
	}
});

var Router = Backbone.Router.extend({
	routes: {
		"": "index",
		"about": "about"
	},
	initialize: function(opts) {
		this.tiles = opts.tiles || new Tiles([]);
		this.app = new App({collection: this.tiles}).render();
		Backbone.history.start({});
	},
	index: function() {
		this.app.activate("index");
	},
	about: function() {
		this.app.activate("about");
	},
});

return {
	Tile: Tile,
	Tiles: Tiles,
	App: App,
	Router: Router,
};

});
