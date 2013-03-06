/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["base", "conn", "view/editor"], function(base, conn, createEditor) {

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
	}
});

var Files = Backbone.Collection.extend({model:File});

var FileListItem = base.ListItemView.extend({
	template: _.template('<a href="#file<%- getpath() %>"><%- get("Name") %></a>'),
});

var FileList = base.ListView.extend({
	itemView: FileListItem,
});

var FileView = Backbone.View.extend({
	tagName: "section",
	className: "file",
	template: _.template('<header><%= crumbs() %></header><%= get("Error") || "" %>'),
	initialize: function(opts) {
		this.model = new File({Path:opts.Path});
		this.content = $('<div class="content">');
		this.editor = null;
		this.line = 0;
		this.children = new Files();
		this.listview = new FileList({collection:this.children});
		this.listenTo(conn, "msg:stat msg:stat.err", this.openMsg);
		this.listenTo(this.model, "change", this.render);
		this.listenTo(this.model, "remove", this.remove);
		conn.send("stat", opts.Path);
	},
	render: function() {
		this.$el.html(this.template(this.model));
		this.$el.append(this.content);
		return this;
	},
	openMsg: function(data) {
		var view = this;
		if (data.Path != this.model.get("Path")) {
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
			var path = data.Path;
			$.get("/raw"+path, function(data){
				view.editor = createEditor(view.content[0], path, data);
				view.editor.commands.addCommands([{
					name:"save",
					bindKey: {win: "Ctrl-S", mac:"Command-S"},
					exec: function(editor, line) {
						view.save();
					},
					readOnly: false
				}]);
				if (view.line > 0) {
					view.setLine(view.line);
				}
			});
		}
	},
	save: function() {
		var path = this.model.get("Path");
		console.log("save", path);
		$.ajax({
			type: "POST",
			url:  "/raw"+path,
			data: this.editor.getSession().getValue(),
			processData: false,
			success: function(resp) {
				console.log("save success", path);
			},
		});
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

var views = {};
function openfile(path) {
	var line = 0;
	var pathline = path.split("#L");
	if (pathline.length > 1 && pathline[1].match(/^\d+$/)) {
		path = pathline[0], line = parseInt(pathline[1], 10);
	}
	if (path && path[path.length-1] == "/") {
		path = path.slice(0, path.length-1);
	}
	path = "/"+path;
	var view = views[path];
	if (!view) {
		view = new FileView({id: _.uniqueId("file"), Path: path});
		views[path] = view;
	}
	var name = _.map(_.last(pathcrumbs(path),2), function(p) {
		return p[1];
	}).join("/") || path;
	if (line > 0) view.setLine(line);
	return {
		id: view.id,
		uri: "file"+path,
		name: name,
		view: view,
		active: true,
		closable: true,
	};
}

return {
	View: FileView,
	router: {
		route:    "file/*path",
		name:     "openfile",
		callback: openfile,
	},
};
});
