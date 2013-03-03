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
			' <span class="close">x</span>',
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
	initialize: function(opts) {
		this.tiles = opts.tiles || new Tiles([]);
		this.app = new App({collection: this.tiles}).render();
		this.tiles.each(function(t) {
			var id = t.get("id"), uri = t.get("uri");
			this.route(uri, id, _.bind(this.app.activate, this.app, id));
		}, this);
		_.each(opts.tilerouters, function(tr) {
			this.route(tr.route, tr.name, _.bind(this.maketile, this, tr));
		}, this)
		Backbone.history.start({});
	},
	maketile: function(tilerouter) {
		var tiles = tilerouter.callback.apply(tilerouter, _.rest(arguments));
		if (!tiles) return
		tiles = _.isArray(tiles) ? tiles : [tiles];
		this.tiles.add(tiles);
		var active = _.find(tiles, function(t) {return t.active;});
		if (active) {
			this.app.activate(active.id);
		}
	}
});

return {
	Tile: Tile,
	Tiles: Tiles,
	App: App,
	Router: Router,
};

});
