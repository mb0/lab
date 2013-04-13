/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["base", "tile"], function(base, tile) {

var Tab = base.ListItemView.extend({
	template: _.template([
		'<a href="#<%- get("uri") %>" <% if (get("active")) { %> class="active" <% } %>>',
		'<%= get("name") %></a>',
		'<% if (get("close")) { %>',
		'<i class="close icon-remove" title="close"></i>',
		'<% } %>',
	].join('')),
	events: { "click .close": "removeModel"},
	removeModel: function() {
		this.collection.remove(this.model);
	}
});

var Tabs = base.ListView.extend({
	itemView: Tab
});

var App = Backbone.View.extend({
	el: $("#app").get(0),
	initialize: function() {
		this.active = null;
		this.history = [];
		this.tabs = new Tabs({collection: this.collection}).render();
		$('<nav>').appendTo(this.$el).append(this.tabs.$el);
		this.$cont = $('<div>').appendTo(this.$el);
		this.listenTo(this.collection, "reset", this.render);
		this.listenTo(this.collection, "remove", this.tileRemoved);
	},
	activate: function(id) {
		if (this.active) {
			if (this.active.get("id") == id) return;
			this.active.set("active", false);
			this.$cont.children().remove();
		}
		this.active = this.collection.get(id);
		if (this.active) {
			this.active.set("active", true);
			this.$cont.append(this.active.get("view").$el);
			this.history = _.without(this.history, id);
			this.history.push(id);
			if (this.history.length > 50) {
				this.history = _.last(this.history, 50);
			}
		}
	},
	render: function() {
		var active = this.collection.find(function(model) {
			return model.get("active");
		});
		if (active !== undefined) {
			this.activate(active.get("id"));
		}
		return this;
	},
	tileRemoved: function(tile) {
		this.history = _.without(this.history, tile.id);
		if (!tile.get("active")) return;
		var last = _.last(this.history);
		if (last !== undefined) {
			var uri = this.collection.get(last).get("uri");
			Backbone.history.navigate(uri, {trigger: true});
		}
	},
});


var Router = Backbone.Router.extend({
	initialize: function(opts) {
		this.tiles = opts.tiles || new tile.Tiles([]);
		this.app = new App({collection: this.tiles}).render();
		this.tiles.each(function(t) {
			var id = t.get("id"), uri = t.get("uri");
			this.route(uri, id, _.bind(this.app.activate, this.app, id));
		}, this);
		_.each(opts.tilerouters, function(tr) {
			this.route(tr.route, tr.name, _.bind(this.makeTile, this, tr));
		}, this);
		Backbone.history.start({});
	},
	makeTile: function(tileRouter) {
		var tiles = tileRouter.tiles.apply(tileRouter, _.rest(arguments));
		if (!tiles) return;
		tiles = _.isArray(tiles) ? tiles : [tiles];
		this.tiles.add(tiles);
		var active = _.find(tiles, function(t) {
			return this.tiles.get(t.id).get("active");
		}, this);
		if (active) {
			this.app.activate(active.id);
		}
	}
});

return {
	App: App,
	Router: Router,
};
});
