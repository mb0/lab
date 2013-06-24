// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

define(["require", "ace/lib/event", "ace/editor", "ace/edit_session", "ace/undomanager", "ace/virtual_renderer", "ace/multi_select"],
function(require, event) {

function createRenderer(container, theme) {
	var R = require("ace/virtual_renderer").VirtualRenderer;
	var r = new R(container, theme || {
		cssClass: "ace_lab",
		isDark:true,
	});
	r.setAnimatedScroll(true);
	r.setShowGutter(true);
	r.setShowPrintMargin(true);
	r.setPrintMarginColumn(120);
	r.setHScrollBarAlwaysVisible(false);
	return r;
}

function createSession(value, mode) {
	var S = require("ace/edit_session").EditSession;
	var U = require("ace/undomanager").UndoManager;
	var s = new S(value, mode);
	s.setUndoManager(new U());
	s.setUseSoftTabs(false);
	s.setTabSize(8);
	return s;
}

function createEditor(renderer, session, multi) {
	var E = require("ace/editor").Editor;
	var M = require("ace/multi_select").MultiSelect;
	var e = new E(renderer, session);
	if (multi) new M(e);
	e.setScrollSpeed(5);
	e.setDragDelay(3000);
	e.setHighlightActiveLine(true);
	e.setHighlightSelectedWord(true);

	var r = e.resize.bind(e, null);
	event.addListener(window, "resize", r);
	e.on("destroy", function() {
		event.removeListener(window, "resize", r);
	});
	return e;
}

function createModes(map) {
	var modes = {};
	for (var key in map) {
		var mode = {name: key, path: "ace/mode/"+ key};
		mode.regex = new RegExp("^.*\\.(" + map[key] + ")$");
		modes[key] = mode;
	}
	return modes;
}

var modes = createModes({
	css: "css",
	golang: "go",
	html: "htm|html|xhtml",
	javascript: "js",
	json: "json",
	markdown: "md|markdown",
	text: "txt",
	xml: "xml|rdf|rss|wsdl|xslt|atom|mathml|mml|xul|xbl",
	c_cpp: "c|cc|cpp|cxx|h|hh|hpp",
	diff: "diff|patch",
	sql: "sql",
	svg: "svg",
	tcl: "tcl",
	toml: "toml",
	yaml: "yaml",
});

function getMode(path) {
	for (var key in modes) {
		var mode = modes[key];
		if (path.match(mode.regex)) {
			return mode;
		}
	}
	return modes.text;
}

return {
	createRenderer: createRenderer,
	createSession:  createSession,
	createEditor:   createEditor,
	modes: modes,
	getMode: getMode,
};
});
