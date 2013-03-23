/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["conn", "view/modes", "view/ace", "view/docs", "lib/paths", "lib/completion"],
function(conn, modes, ace, docs, paths, completion) {

function getCommands(doc) {
	var list = [{
		name: "save", readOnly: false,
		exec: function() {
			doc.publish();
		},
		bindKey: {win: "Ctrl-S", mac:"Command-S"},
	}];
	if (!doc.get("Path").match(/\.go$/)) {
		return list;
	}
	list.push({
		name: "complete", readOnly: false,
		exec: function(editor) {
			doc.complete(editor.selection.getCursor());
		},
		bindKey: {win: "Ctrl-Space", mac:"Command-Space"},
	});
	list.push({
		name: "format", readOnly: false,
		exec: function() {
			doc.format(doc);
		},
		bindKey: {win: "Ctrl-Shift-F", mac:"Command-Shift-F"},
	});
	return list;
}

var File = Backbone.Model.extend({
	idAttribute: "Id",
	getPath: function() {
		return this.get("Path");
	},
	getIcon: function() {
		return 'icon-file';
	},
	getCrumbs: function() {
		return _.map(paths.crumbs(this.getPath()), function(c) {
			return '<a href="#file/'+ c[0] +'">/'+c[1]+"</a>";
		}).join("");
	},
});

var EditorView = Backbone.View.extend({
	template: _.template(
		'<header><i class="<%- getIcon() %>"></i> <%= getCrumbs() %></header>'
	),
	initialize: function() {
		this.$el.addClass("editor");
		this.model = this.model.isValid ? this.model : new File(this.model);
		this.$el.html(this.template(this.model));
		this.$editor = $('<div class="content">').appendTo(this.$el);
		this.editor = null;
		this.doc = docs.subscribe(this.model.id, this.model.getPath());
		this.listenTo(conn, "msg:complete", this.onMsgComplete);
		this.listenTo(this.doc, "change:Ace", this.onChangeAce);
		this.listenTo(this.model, "remove", this.remove);
	},
	render: function() {
		this.$("> header").replaceWith(this.template(this.model));
		return this;
	},
	onChangeAce: function() {
		var mode = modes.matchPath(this.model.getPath());
		var renderer = ace.createRenderer(this.$editor.get(0));
		var session = ace.createSession(this.doc.get("Ace"), mode.get("mode"));
		this.editor = ace.createEditor(renderer, session);
		this.editor.commands.addCommands(getCommands(this.doc));
		if (this.line > 0) {
			this.setLine(this.line);
		}
		if (mode.id !== "golang") {
			return;
		}
		this.editor.on("click", function(e) {
			if (!e.getAccelKey() || !e.domEvent.altKey) return;
			var pos = e.getDocumentPosition();
			var sess = e.editor.getSession();
			var str = sess.getTokenAt(pos.row, pos.column);
			if (str.type !== "string") return;
			pos.column = str.index-1;
			var skip = ["string", "text", "paren.lparen", "identifier"];
			var tok = ace.previousToken(sess, pos, "keyword", skip);
			if (!tok || tok.value !== "import") return;
			var path = str.value.substr(1, str.value.length-2);
			Backbone.history.navigate("doc/"+ path, {trigger: true});
		});
	},
	onMsgComplete: function(data) {
		if (data.Id !== this.model.id) return;
		completion.show(this.editor, data);
	},
	setLine: function(l) {
		if (this.editor !== null) {
			this.editor.moveCursorToPosition({row:l-1, column:0});
			var row = l-(this.editor.$getVisibleRowCount()*0.5);
			this.editor.scrollToRow(Math.max(row,0));
		} else {
			this.line = l;
		}
	},
});

return {
	View: EditorView,
};
});

