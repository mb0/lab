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
		var view = this;
		$.get("/doc/"+opts.Path, function(data){
			data = data.replace(/ href="#/g, ' data-href="#').replace(/ id="/g, ' data-id="');
			view.$el.html(data);
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
		if (href.indexOf("#file/") == 0) {
			Backbone.history.navigate(href, {trigger: true});
			return;
		}
		var target = $('[data-id="' + href.slice(1)+'"]');
		if (target.hasClass('toggle')) {
			target.toggleClass('toggle').toggleClass('toggleVisible');
		}
		target.get(0).scrollIntoView();
		var hashes = location.hash.split('#');
		Backbone.history.navigate('#'+hashes[1]+href);
	}
});

var views = {};
function opendoc(path) {
	if (path && path[path.length-1] == "/") {
		path = path.slice(0, path.length-1);
	}
	var view = views[path];
	if (!view) {
		view = new DocView({id: _.uniqueId("doc"), Path: path});
		views[path] = view;
	}
	return {
		id: view.id,
		uri: "doc/"+path,
		name: path,
		view: view,
		active: true,
		closable: true,
	};
}

return {
	View: DocView,
	router: {
		route:    "doc/*path",
		name:     "opendoc",
		callback: opendoc,
	},
};
});
