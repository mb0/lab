/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["base", "conn", "lib/sotdoc"],
function(base, conn, sotdoc) {

var icons = {
	subscribe: "icon-cloud-download",
	waiting:   "icon-cloud-upload",
	received:  "icon-cloud-download",
	published: "icon-hdd",
};

function getIcon(name, defaultIcon) {
	var icon = icons[name] || defaultIcon;
	if (!icon) return '';
	return '<i class="'+ icon +'"></i>';
}

var Doc = sotdoc.Doc.extend({
	icon: function() {
		return getIcon(this.get("Status"), "icon-cloud");
	},
	publish: function() {
		conn.send("publish", {Id: this.get("Id")});
	},
	format: function() {
		conn.send("format", {Id: this.get("Id")});
	},
	complete: function(cursor) {
		var acedoc = this.get("Ace");
		var lines = acedoc.$lines || acedoc.getAllLines();
		var idxr = sotdoc.posToRestIndex(lines, cursor);
		conn.send("complete", {Id: this.get("Id"), Offs: idxr.start});
	}
});

var Docs = Backbone.Collection.extend({model:Doc});

var DocListItem = base.ListItemView.extend({
	template: _.template('<a href="#file<%- get("Path") %>"><%= icon() %> #<%- get("Rev") %> <%- get("Path") %></a>'),
});

var DocList = base.ListView.extend({
	itemView: DocListItem,
});

var DocsView = Backbone.View.extend({
	tagName: "section",
	className: "docs",
	template: _.template('<header><i class="icon-inbox"></i> Documents</header>'),
	initialize: function(opts) {
		this.listview = new DocList({collection:this.collection});
		this.listenTo(conn, "msg:subscribe", this.onSubscribe);
		this.listenTo(conn, "msg:revise", this.onRevise);
		this.listenTo(conn, "msg:publish", this.onPublish);
		this.listenTo(conn, "msg:unsubscribe", this.onUnsubscribe);
		this.render();
	},
	render: function() {
		this.$el.html(this.template(this.model));
		this.$el.append(this.listview.el);
		return this;
	},
	onSubscribe: function(data) {
		var doc = this.collection.get(data.Id);
		if (!doc) {
			console.log("subscribe unknown document", data);
			return;
		}
		var text = data.Ops && data.Ops[0] || "";
		doc.createAce(data.Rev, data.User, text);
		doc.on("ops", function(doc, ops) {
			conn.send("revise", {Id: doc.id, Rev: doc.get("Rev"), Ops: ops});
		});
	},
	onPublish: function(data) {
		var doc = this.collection.get(data.Id);
		if (!doc) {
			console.log("publish unknown document", data);
			return;
		}
		doc.set("Status", "published");
	},
	onRevise: function(data) {
		var doc = this.collection.get(data.Id);
		if (!doc) {
			console.log("revise unknown document", data);
			return;
		}
		var err = null;
		if (doc.get("User") === data.User) {
			err = doc.ackOps(data.Ops);
		} else {
			err = doc.recvOps(data.Ops);
		}
		if (err !== null) {
			console.log("revise error", err);
		}
	},
	onUnsubscribe: function(data) {
		var doc = this.collection.get(data.Id);
		if (!doc) {
			console.log("unsubscribe unknown document", data);
			return;
		}
		this.collection.remove(doc);
	}
});
var docs = new Docs();
var view = new DocsView({collection:docs});
return {
	view: view,
	subscribe: function(id, path) {
		var doc = new Doc({Id: id, Path: path, Rev: -1, Status: "subscribe"});
		docs.add(doc);
		conn.send("subscribe", {Id: id});
		return doc;
	},
};
});
