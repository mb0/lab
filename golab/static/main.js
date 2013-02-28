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
	getoutput: function() {
		var out = (this.get("Stdout") || "") + (this.get("Stderr") || "");
		return out.replace(/(^(#.*|\S)\n|\n#[^\n]*)/g, "");
	},
	getstatus: function() {
		return !this.get("Err") ? "ok" : "fail";
	},
});
var Reports = Backbone.Collection.extend({model:Report});
var ReportListItem = base.ListItemView.extend({
	template1: _.template([
		'<div class="report <%- getstatus() %>">',
		'<header>',
		'<span class="status"><%- getstatus().toUpperCase() %></span> ',
		'<span class="mode"><%= get("Mode") %></span> ',
		'<%= get("Path") %> <%= get("Err") || "" %>',
		'</header>',
		'<% var o; if (o = getoutput()) { %><pre><%= o %></pre><% } %>',
		'</div>',
	].join('')),
	template: function(r) {
		console.log("addreport", r);
		return this.template1(r);
	},
});
var ReportList = base.ListView.extend({
	itemView: ReportListItem
});
var ReportView = base.Page.extend({
	initialize: function() {
		this.reports = new Reports();
		this.listview = new ReportList({collection:this.reports});
		this.listenTo(conn, "msg:report", this.addReport);
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
		' * json2.js (public domain).',
		'</pre>'
	].join('\n'))},
])});
$("#app > nav > ul").prepend("<li>golab</li>")

conn.connect();

});
