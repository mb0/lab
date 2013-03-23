/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["conn", "tile", "view/folder", "view/editor", "view/report", "lib/paths"],
function(conn, tile, folder, editor, report, paths) {

var FileView = Backbone.View.extend({
	tagName: "section",
	className: "file",
	initialize: function(opts) {
		this.id = _.uniqueId("file");
		this.path = opts.path;
		this.line = opts.line;
		this.tile = new tile.Tile({
			id:     this.id,
			uri:    "file"+this.path,
			name:   paths.shorten(this.path, 2),
			view:   this,
			active: true,
			close:  true,
		});
		this.content = null;
		this.listenTo(this.tile, "remove", this.remove);
		this.listenTo(conn, "msg:stat msg:stat.err", this.onstat);
		conn.send("stat", this.path);
	},
	render: function() {
		return this;
	},
	onstat: function(data) {
		if (data.Path != this.path) return;
		var opts = {el: this.el, model: data, tile: this.tile};
		if (data.Error)  {
			alert(data.Error);
		} else if (data.IsDir || data.Children) {
			this.content = new folder.View(opts);
		} else {
			this.content = new editor.View(opts);
			if (this.line > 0) {
				this.content.setLine(this.line);
			}
		}
	},
	isEditor: function() {
		return this.content && this.content.$el && this.content.$el.hasClass("editor");
	},
});

var FileRouter = Backbone.View.extend({
	initialize: function(opts) {
		this.map = {}; // path: {view, annotations},
		this.listenTo(report.view.reports, "add change reset", this.checkreports);
		this.route = "file/*path";
		this.name = "openfile";
	},
	checkReports: function() {
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
		_.delay(this.updateAnnotations, 200, update);
	},
	updateAnnotations: function(work) {
		_.each(work, function(e) {
			if (!e.view || !e.view.isEditor()) return;
			var editor = e.view.content.editor;
			if (editor) editor.session.setAnnotations(e.markers);
		});
	},
	callback: function(path) {
		var pl = paths.splitHashLine("/"+path);
		path = pl.path;
		var entry = this.map[path] || {view: null, markers: []};
		if (!entry.view) {
			entry.view = new FileView(pl);
			this.map[path] = entry;
		} else if (pl.line > 0 && entry.view.isEditor()) {
			entry.view.content.setLine(pl.line);
		}
		return entry.view.tile;
	},
});

return {
	router: new FileRouter(),
};
});

