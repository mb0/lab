/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
require.config({
	paths: {
		json2: 'libs/json2.min',
		zepto: 'libs/zepto.min',
		underscore: 'libs/underscore',
		backbone: 'libs/backbone'
	},
	shim: {
		underscore: {exports: "_"},
		zepto: {exports: "$"},
		backbone: {exports: "Backbone", deps: ["underscore", "zepto"]},
	}
});

define(["conn", "app", "view/report", "view/file"], function(conn, app, report, file) {

$(document).on("click", "a", function(e) {
	e.preventDefault();
	Backbone.history.navigate(e.target.getAttribute("href"), {trigger: true});
});

var Html = Backbone.View.extend({
	tagName: "div",
	constructor: function(text, opts) {
		this.text = text;
		Backbone.View.prototype.constructor.call(this, opts);
		this.render();
	},
	render: function() {
		this.$el.html(this.text);
		return this;
	},
});

new app.Router({
	tilerouters: [file.router],
	tiles: new app.Tiles([
		{id: "index", uri: "", name:"report", view: new report.View()},
		{id: "about", uri: "about", name:"about", view: new Html([
			'<pre>',
			'go live action builds',
			'=====================\n',
			'&copy; 2013 Martin Schnabel. All rights reserved.',
			'BSD-style license.\n',
			'Other code used:',
			' * github.com/garyburd/go-websocket (Apache License 2.0)',
			' * Underscore, Zepto.js, Backbone.js (MIT License)',
			' * require.js (BSD/MIT License)',
			' * json2.js (public domain).',
			'</pre>'
		].join('\n'))},
	])
});

var ConnView = Backbone.View.extend({
	tagName: "li",
	events: {
		"click a": "connect",
	},
	initialize: function() {
		this.listenTo(conn, "open close", this.render);
		this.render();
	},
	render: function() {
		this.$el.html(conn.connected() ? 'golab' : '<a class="offline">connect</a>');
		return this;
	},
	connect: function(e) {
		e.preventDefault();
		conn.connect();
	}
});

$("#app > nav > ul").prepend(new ConnView({}).render().el)

conn.connect();

});
