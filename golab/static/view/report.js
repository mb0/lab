/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["base", "conn"], function(base, conn) {

var Report = Backbone.Model.extend({
	idAttribute: "Id",
	getresult: function() {
		var src = this.get("Src").Result, test = this.get("Test").Result;
		if (src && src.Err) return src;
		if (test) return test;
		return src;
	},
	haserrors: function(res) {
		res = res || this.getresult();
		return res && res.Err != null;
	},
	getoutput: function(res) {
		if (!res) return "";
		return (res.Stdout || "") + (res.Stderr || "");
	},
	fixoutput: function(out) {
		out = out.replace(/(\/([^\/\s]+\/)+(\S+?\.go))\:(\d+)(?:\:(\d+))?\:/g, '<a href="#file$1#L$4">$2$3:$4</a>');
		return out.replace(/(^(#.*|\S)\n|\n#[^\n]*)/g, "");
	},
	getfiles: function() {
		var res, files = [];
		if ((res = this.get("Src")) && res.Info)
			files = files.concat(res.Info.Files);
		if ((res = this.get("Test")) && res.Info)
			files = files.concat(res.Info.Files);
		return files;
	}
});

var Reports = Backbone.Collection.extend({model:Report});

var ReportListItem = base.ListItemView.extend({
	template: _.template([
		'<% var res = getresult(); var err = haserrors(res), o = getoutput(res) %>',
		'<div class="report <%- err ? "fail" : "ok" %>">',
		'<header>',
		'<span class="status"><%- err ? "FAIL" : "OK" %></span> ',
		'<span class="mode"><%= res && res.Mode || "" %></span> ',
		'<a href="#file<%= get("Dir") %>"><%= get("Path") %></a> <%= res && res.Err || "" %>',
		'</header>',
		'<% if (o) { %><pre><%= fixoutput(o) %></pre><% } %>',
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
		this.lookup = {};
		this.reports = new Reports();
		this.listview = new ReportList({collection:this.reports});
		this.listenTo(conn, "msg:report msg:reports", this.addreports);
		this.render();
	},
	render: function() {
		this.$el.append(this.listview.render().$el);
		return this;
	},
	addreports: function(data) {
		if (!_.isArray(data)) data = [data];
		this.reports.add(data, {merge: true});
		this.scrolltolast();
		var hasErrors = this.reports.find(function(r) {
			return r.haserrors();
		});
		var nav = $('i[title="report"]');
		if (hasErrors) nav.addClass('red').removeClass('green');
		else nav.addClass('green').removeClass('red');
	},
	scrolltolast: function() {
		var d = this.el.scrollHeight - this.el.clientHeight;
		if (d > 0) this.el.scrollTop = d;
	}
});

return {view: new ReportView()};
});
