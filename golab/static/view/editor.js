/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["require", "ace/ace", "ace/editor", "ace/virtual_renderer", "ace/multi_select"], function(require, ace) {

var Renderer = require("ace/virtual_renderer").VirtualRenderer,
    Editor = require("ace/editor").Editor,
    MultiSelect = require("ace/multi_select").MultiSelect;

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
			return modes[i].mode;
		}
	}
	return modesByName.text;
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
	sess.setMode(getMode(name));
	sess.setUseSoftTabs(false);
	sess.setTabSize(8);
	var editor = new Editor(renderer);
	editor.setScrollSpeed(5);
	editor.setDragDelay(3000);
	new MultiSelect(editor);
	editor.setSession(sess);
	editor.setReadOnly(true);
	editor.setHighlightActiveLine(true);
	editor.setHighlightSelectedWord(true);
	return editor;
};
});
