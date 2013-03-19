/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["base", "conn", "view/editor", "view/report", "view/sotdoc", "ace/document"],
function(base, conn, createEditor, report, sotdoc, acedoc) {

function pathcrumbs(path) {
	if (!path) return [];
	var i = 0;
	if (path[0] == "/") path = path.substr(1);
	return _.map(path.split("/"), function(p){
		i += p.length;
		return [path.substr(0, i++), p];
	});
}

var File = Backbone.Model.extend({
	idAttribute: "Id",
	getpath: function() {
		var path = this.get("parent").get("Path");
		if (path && path[path.length-1] != "/") {
			path += "/";
		}
		return path + this.get("Name");
	},
	crumbs: function() {
		return _.map(pathcrumbs(this.get("Path")), function(c) {
			return '<a href="#file/'+ c[0] +'">/'+c[1]+"</a>";
		}).join("");
	},
	icon: function(open) {
		if (!this.get("IsDir")) {
			return 'icon-file';
		} else if (open) {
			return 'icon-folder-open-alt';
		}
		return 'icon-folder-close-alt';
	}
});

var Files = Backbone.Collection.extend({model:File});

var FileListItem = base.ListItemView.extend({
	template: _.template('<a href="#file<%- getpath() %>"><i class="<%- icon() %>"></i> <%- get("Name") %></a>'),
});

var FileList = base.ListView.extend({
	itemView: FileListItem,
});
var FileView = Backbone.View.extend({
	tagName: "section",
	className: "file",
	template: _.template('<header><i class="<%- icon(true) %>"></i> <%= crumbs() %></header><%= get("Error") || "" %>'),
	initialize: function(opts) {
		this.model = new File({Path:opts.Path});
		this.content = $('<div class="content">');
		this.editor = null;
		this.doc = null;
		this.line = 0;
		this.children = new Files();
		this.listview = new FileList({collection:this.children});
		this.listenTo(conn, "msg:stat msg:stat.err", this.onstat);
		this.listenTo(this.model, "change", this.render);
		this.listenTo(this.model, "remove", this.remove);
		conn.send("stat", opts.Path);
	},
	render: function() {
		this.$el.html(this.template(this.model));
		this.$el.append(this.content);
		return this;
	},
	onstat: function(data) {
		var view = this;
		var path = this.model.get("Path")
		if (data.Path != path) {
			return;
		}
		this.model.set(data);
		if (data.Error) return;
		if (data.IsDir || data.Children) {
			_.each(data.Children, function(c){c.parent = view.model;});
			this.children.reset(data.Children);
			this.content.children().remove();
			this.content.append(this.listview.$el);
		} else {
			this.listenTo(conn, "msg:subscribe", this.onsub);
			this.listenTo(conn, "msg:publish", this.onpub);
			this.listenTo(conn, "msg:revise", this.onrev);
			conn.send("subscribe", {Id: data.Id});
		}
	},
	onsub: function(data) {
		if (data.Id != this.model.get("Id")) return;
		var text = data.Ops && data.Ops[0] || "";
		this.doc = new acedoc.Document(text);
		this.doc.user = data.User;
		var sdoc = sotdoc.install(data.Id, data.Rev, this.doc);
		var path = this.model.get("Path");
		this.editor = createEditor(this.content[0], path, this.doc);
		this.doc.on("sotops", function(e) {
			conn.send("revise", e.data);
		});
		this.editor.commands.addCommands([{
			name:"save",
			bindKey: {win: "Ctrl-S", mac:"Command-S"},
			exec: function(editor, line) {
				console.log("publish", path);
				conn.send("publish", {Id: sdoc.id});
			},
			readOnly: false
		}]);
		if (this.line > 0) {
			this.setLine(this.line);
		}
	},
	onpub: function(data) {
		if (this.doc && data.Id != this.doc.sotdoc.id) return;
		console.log("saved", this.model.get("Path"));
	},
	onrev: function(data) {
		if (this.doc && data.Id != this.doc.sotdoc.id) return;
		var err = null;
		if (this.doc.user && this.doc.user == data.User) {
			err = sotdoc.ackOps(this.doc, data.Ops);
		} else {
			err = sotdoc.recvOps(this.doc, data.Ops);
		}
	},
	setLine: function(l) {
		if (this.editor != null) {
			this.editor.moveCursorToPosition({row:l-1, column:0});
			var row = l-(this.editor.$getVisibleRowCount()*0.5);
			this.editor.scrollToRow(Math.max(row,0));
		} else {
			this.line = l;
		}
	}
});

var ViewManager = Backbone.View.extend({
	initialize: function(opts) {
		this.map = {}; // path: {view, annotations},
		this.listenTo(report.view.reports, "add change reset", this.checkreports);
		this.route = "file/*path";
		this.name = "openfile";
	},
	checkreports: function() {
		var reports = report.view.reports;
		// check for markers and add to map
		var re = /^((\/(?:[^\/\s]+\/)+)?(\S+?\.go))\:(\d+)\:(?:(\d+)\:)?(.*)$/;
		var update = {}, path, entry;
		reports.each(function(e) {
			var res = e.getresult();
			_.each(e.getfiles(), function(file) {
				path = e.get("Dir")+"/"+file.Name;
				if ((entry = this.map[path])) {
					entry.markers = [];
					update[path] = entry;
				}
			}, this);
			if (!e.haserrors(res)) {
				return;
			}
			var out = e.getoutput(res).split("\n");
			var line, m;
			for (var i = 0; i < out.length; i++) {
				m = out[i].match(re);
				if (!m) continue;
				path = m[2] ? m[1] : e.get("Dir")+ "/"+ m[3];
				line = parseInt(m[4], 10);
				entry = this.map[path] || {view: null, markers: []};
				entry.markers.push({
					row: line - 1,
					column: (m[5] ? parseInt(m[5], 10) : 0) -1,
					text: m[6].trim(),
					type: "error"
				});
				this.map[path] = entry;
				update[path] = entry;
			}
		}, this);
		_.delay(this.updateannotations, 200, update);
	},
	updateannotations: function(work) {
		_.each(work, function(e) {
			if (!e.view || !e.view.editor) return;
			e.view.editor.getSession().setAnnotations(e.markers);
		});
	},
	callback: function(path) {
		var pl = this.splitline(path);
		path = pl[0];
		var entry = this.map[path] || {view: null, markers: []};
		if (!entry.view) {
			entry.view = new FileView({id: _.uniqueId("file"), Path: path});
			this.map[path] = entry;
		}
		if (pl[1] > 0) entry.view.setLine(pl[1]);
		return this.newtile(path, entry.view);
	},
	newtile: function(path, view) {
		return {
			id: view.id,
			uri: "file"+path,
			name: this.tilename(path),
			view: view,
			active: true,
			closable: true,
		};
	},
	tilename: function(path) {
		return _.map(_.last(pathcrumbs(path),2), function(p) {
			return p[1];
		}).join("/") || path;
	},
	splitline: function(path) {
		var line = 0;
		var pathline = path.split("#L");
		if (pathline.length > 1 && pathline[1].match(/^\d+$/)) {
			path = pathline[0], line = parseInt(pathline[1], 10);
		}
		if (path && path[path.length-1] == "/") {
			path = path.slice(0, path.length-1);
		}
		return ["/"+path, line];
	}
});

return {
	View: FileView,
	router: new ViewManager(),
};
});
