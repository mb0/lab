/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["require", "ace/ace", "ace/editor", "ace/virtual_renderer", "ace/multi_select", "ace/lib/event"],
function(require, ace) {

var	Editor = require("ace/editor").Editor,
	Renderer = require("ace/virtual_renderer").VirtualRenderer,
	MultiSelect = require("ace/multi_select").MultiSelect,
	event = require("ace/lib/event");

var modesByName = {
	css: "css",
	golang: "go",
	html: "htm|html|xhtml",
	javascript: "js",
	json: "json",
	markdown: "md|markdown",
	text: "txt",
	xml: "xml|rdf|rss|wsdl|xslt|atom|mathml|mml|xul|xbl"
};

function Mode(name, extensions) {
	this.name = name;
	this.mode = "ace/mode/" + name;
	this.extRe = new RegExp("^.*\\.(" + extensions + ")$");
}

var modes = (function (modes){
	for (var name in modesByName) {
		var mode = new Mode(name, modesByName[name]);
		modesByName[name] = mode;
		modes.push(mode);
	}
	return modes;
})([]);

function getMode(path) {
	for (var i = 0; i < modes.length; i++) {
		if (path.match(modes[i].extRe)) {
			return modes[i];
		}
	}
	return modesByName.text;
}

function getLastToken(sess, r, i, type, ignore) {
	for (var toks = null; i >= 0 || r > 0; i--) {
		if (toks === null || i < 0) {
			toks = sess.getTokens(i < 0 ? --r : r);
			i = toks.length -1;
		}
		if (i < 0) continue;
		var tt = toks[i].type;
		if (tt == type) return toks[i];
		if (_.isArray(ignore) &&  _.contains(ignore, tt)) continue;
		break;
	}
}

return function(c, name, value) {
	if (c.env && c.env.editor instanceof Editor)
		return c.env.editor;
	var renderer = new Renderer(c, {isDark:true, cssClass: "ace_lab"});
	renderer.setAnimatedScroll(true);
	renderer.setShowGutter(true);
	renderer.setShowPrintMargin(true);
	renderer.setPrintMarginColumn(120);
	renderer.setHScrollBarAlwaysVisible(false);
	var sess = ace.createEditSession(value);
	var mode = getMode(name);
	sess.setMode(mode.mode);
	sess.setUseSoftTabs(false);
	sess.setTabSize(8);
	var editor = new Editor(renderer);
	editor.setScrollSpeed(5);
	editor.setDragDelay(3000);
	new MultiSelect(editor);
	editor.setSession(sess);
	editor.setReadOnly(false);
	editor.setHighlightActiveLine(true);
	editor.setHighlightSelectedWord(true);
	var env = {
		document: sess,
		editor: editor,
		onResize: editor.resize.bind(editor, null)
	};
	event.addListener(window, "resize", env.onResize);
	editor.on("destroy", function() {
		event.removeListener(window, "resize", env.onResize);
	});
	if (mode.name == "golang") {
		editor.on("click", function(e) {
			if (!e.getAccelKey() || !e.domEvent.altKey) return;
			var pos = e.getDocumentPosition();
			var sess = e.editor.getSession();
			var tok = sess.getTokenAt(pos.row, pos.column);
			if (tok.type == "string") {
				var kw = getLastToken(sess, pos.row, tok.index-1, "keyword", ["string", "text", "paren.lparen", "identifier"]);
				if (kw && kw.value == "import") {
					var path = tok.value.substr(1, tok.value.length-2);
					Backbone.history.navigate("doc/"+ path, {trigger: true});
				}
			}
		});
	}
	c.env = editor.env = env;
	return editor;
};
});
