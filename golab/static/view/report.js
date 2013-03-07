/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["base", "conn"], function(base, conn) {

var Report = Backbone.Model.extend({
	getresult: function() {
		var src = this.get("Src").Result, test = this.get("Test").Result;
		if (src && src.Err) return src;
		if (test) return test;
		return src;
	},
	getstatus: function(res) {
		return !(res && res.Err) ? "ok" : "fail";
	},
	getoutput: function(res) {
		if (!res) return "";
		var out = (res.Stdout || "") + (res.Stderr || "");
		out = out.replace(/(\/([^\/\s]+\/)+(\S+?\.go))\:(\d+)(?:\:(\d+))?\:/g, '<a href="#file$1#L$4">$2$3:$4</a>');
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
		'<a href="#file<%= get("Dir") %>"><%= get("Path") %></a> <%= res && res.Err || "" %>',
		'</header>',
		'<% if (o) { %><pre><%= o %></pre><% } %>',
		'</div>',
	].join('')),
});

var ReportList = base.ListView.extend({
	itemView: ReportListItem
});

var ReportView = Backbone.View.extend({
	tagName: "section",
	attributes: {
		"class": "reportview",
	},
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
		var d = this.el.scrollHeight - this.el.clientHeight;
		if (d > 0)
			this.el.scrollTop = d;
	}
});

return {View: ReportView};
});
