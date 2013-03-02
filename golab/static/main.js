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

define(["base", "app", "conn"], function(base, app, conn) {

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
var Report = Backbone.Model.extend({
	getresult: function() {
		var src = this.get("Src").Result, test = this.get("Test").Result;
		if (src && src.Err) return src;
		if (test) return test;
		return src
	},
	getstatus: function(res) {
		return !(res && res.Err) ? "ok" : "fail";
	},
	getoutput: function(res) {
		if (!res) return "";
		var out = (res.Stdout || "") + (res.Stderr || "");
		return out.replace(/(^(#.*|\S)\n|\n#[^\n]*)/g, "");
	},
});
var Reports = Backbone.Collection.extend({model:Report});
var ReportListItem = base.ListItemView.extend({
	template: _.template([
		'<% var res = getresult(); var stat = getstatus(res), o = getoutput(res) %>',
		'<div class="report <%- stat %>">',
		'<header>',
		'<span class="status"><%- stat.toUpperCase() %></span> ',
		'<span class="mode"><%= res && res.Mode || "" %></span> ',
		'<%= get("Path") %> <%= res && res.Err || "" %>',
		'</header>',
		'<% if (o) { %><pre><%= o %></pre><% } %>',
		'</div>',
	].join('')),
});
var ReportList = base.ListView.extend({
	itemView: ReportListItem
});
var ReportView = base.Page.extend({
	initialize: function() {
		this.reports = new Reports();
		this.listview = new ReportList({collection:this.reports});
		this.listenTo(conn, "msg:report msg:reports", this.addReport);
		this.render();
	},
	render: function() {
		this.$el.append(this.listview.render().$el);
		return this;
	},
	addReport: function(data) {
		this.reports.add(data);
	}
});

new app.Router({tiles: new app.Tiles([
	{id: "index", uri: "", name:"reports", view: new ReportView()},
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
])});


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
