/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["backbone"], function() {

var DocView = Backbone.View.extend({
	tagName: "section",
	className: "godoc",
	events: {
		"click .toggleButton": "toggle",
		"click a[data-href]": "toggleLink",
	},
	initialize: function(opts) {
		this.hash = "";
		var view = this;
		var hashes = location.hash.split('#');
		$.get("/doc/"+opts.Path, function(data){
			data = data
				.replace(/ href="#/g, ' data-href="#')
				.replace(/ href="/g, ' href="#'+hashes[1]+'/')
				.replace(/ id="/g, ' data-id="');
			view.$el.html(data);
			var hash = view.hash;
			if (hash) {
				view.hash = null;
				view.openHash(hash);
			}
		});
	},
	render: function() {
		return this;
	},
	toggle: function(e) {
		var el = $(e.target).closest('.toggle, .toggleVisible');
		el.toggleClass('toggle').toggleClass('toggleVisible');
	},
	toggleLink: function(e) {
		e.preventDefault();
		var href = $(e.currentTarget).attr('data-href');
		if (href.indexOf("#file/") === 0) {
			Backbone.history.navigate(href, {trigger: true});
			return;
		}
		this.openHash(href);
	},
	openHash: function(hash) {
		if (!hash) return;
		if (hash[0] == "#") hash = hash.slice(1);
		var target = $('[data-id="'+hash+'"]');
		console.log("found", target);
		if (!target.length) {
			this.hash = hash;
			return;
		}
		if (target.hasClass('toggle')) {
			target.toggleClass('toggle').toggleClass('toggleVisible');
		}
		target.get(0).scrollIntoView();
		var hashes = location.hash.split('#');
		Backbone.history.navigate('#'+hashes[1]+'#'+hash);
		this.hash = "";
	}
});

var ViewManager = Backbone.View.extend({
	initialize: function(opts) {
		this.map = {}; // path: view,
		this.route = "doc/*path";
		this.name = "openfile";
	},
	callback: function(path) {
		var ph = this.splithash(path);
		path = ph[0];
		var view = this.map[path];
		if (!view) {
			view = new DocView({id: _.uniqueId("doc"), Path: path});
			this.map[path] = view;
		}
		if (ph.length > 1 && ph[1]) view.openHash(ph[1]);
		return this.newtile(path, view);
	},
	splithash: function(path) {
		var pathhash = path.split("#");
		if (pathhash.length > 0 && pathhash[0][pathhash[0].length-1] == "/") {
			pathhash[0] = pathhash[0].slice(0, pathhash[0].length-1);
		}
		return pathhash;
	},
	newtile: function(path, view) {
		return {
			id: view.id,
			uri: "doc/"+path,
			name: path,
			view: view,
			active: true,
			closable: true,
		};
	},
});

return {
	View: DocView,
	router: new ViewManager(),
};
});
