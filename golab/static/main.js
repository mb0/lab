/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
require.config({
	paths: {
		json2: 'http://cdnjs.cloudflare.com/ajax/libs/json2/20121008/json2.min',
		zepto: 'http://cdnjs.cloudflare.com/ajax/libs/zepto/1.0rc1/zepto.min',
		underscore: 'http://cdnjs.cloudflare.com/ajax/libs/underscore.js/1.4.4/underscore-min',
		backbone: 'http://cdnjs.cloudflare.com/ajax/libs/backbone.js/0.9.10/backbone-min',
		ace: '/static/ace',
	},
	shim: {
		underscore: {exports: "_"},
		zepto: {exports: "$"},
		backbone: {exports: "Backbone", deps: ["underscore", "zepto"]},
	}
});

define(["conn", "app", "view/report", "view/file", "view/godoc", "unity"], function(conn, app, report, file, godoc) {

$('<link>').attr({
	type: "image/png",
	rel:  "icon",
	href: "/static/golab.png",
}).appendTo($("head").first());

$('<link>').attr({
	type: "text/css",
	rel:  "stylesheet",
	href: "http://cdnjs.cloudflare.com/ajax/libs/font-awesome/3.0.2/css/font-awesome.min.css",
}).appendTo($("head").first());

$(document).on("click", "a", function(e) {
	var href = $(e.currentTarget).attr("href");
	if (href && href.indexOf("http") != 0) {
		e.preventDefault();
		Backbone.history.navigate(href, {trigger: true});
	}
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
	tilerouters: [file.router, godoc.router],
	tiles: new app.Tiles([
		{id: "about", uri: "about", name:'<i class="icon-beaker" title="about"></i>', view: new Html([
			'<pre>',
			'<h3>golab</h3>'+
			'<a href="https://github.com/mb0/lab">github.com/mb0/lab</a> (c) Martin Schnabel '+
			'<a href="https://raw.github.com/mb0/lab/master/LICENSE">BSD License</a>',
			'</pre>'
		].join('\n'))},
		{id: "index", uri: "", name:'<i class="icon-circle" title="report"/></i>', view: report.view},
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
		this.$el.html(!conn.connected() ? '<a class="offline"><i class="icon-signin" title="reconnect"/></i></a>' : '');
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
