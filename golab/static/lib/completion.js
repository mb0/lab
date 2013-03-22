/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
// package completion provides completion popup for the ace editor.
define(["require", "backbone", "ace/editor", "ace/virtual_renderer", "ace/keyboard/hash_handler"],
function(require) {

var Editor = require("ace/editor").Editor;
var Renderer = require("ace/virtual_renderer").VirtualRenderer;
var HashHandler = require("ace/keyboard/hash_handler").HashHandler;

function listEditor(el) {
	var renderer = new Renderer(el, {isDark:true, cssClass: "ace_lab"});
	var editor = new Editor(renderer);
	renderer.setShowGutter(false);
	renderer.setHighlightGutterLine(false);
	renderer.setShowPrintMargin(true);
	renderer.setHScrollBarAlwaysVisible(false);
	renderer.hideCursor();
	editor.setHighlightActiveLine(true);
	return editor;
}

var Proposal = Backbone.Model.extend({
	// class, name, type
});
var Proposals = Backbone.Collection.extend({
	model: Proposal,
});

function localToGlobal(el, pos) {
	var doc = document.documentElement;
	var box = el.getBoundingClientRect();
	return {
		left: pos.left + box.left + window.pageXOffset - doc.clientLeft,
		top:  pos.top  + box.top  + window.pageYOffset - doc.clientTop,
	};
}
function getScreenPixels(editor, at) {
	var r = editor.renderer;
	var pos = r.$cursorLayer.getPixelPosition(at, true);
	pos.left += r.scroller.offsetLeft + r.content.offsetLeft;
	pos = localToGlobal(editor.container, pos);
	return pos;
}

var Popup = Backbone.View.extend({
	tagName: "div",
	id: "completion-popup",
	events: {
		"click": "detach",
	},
	initialize: function(opts) {
		this.proposals = new Proposals();
		this.proposals.on("reset", this.render, this);
		this.$el.hide().appendTo(document.body);
		this.$list = $("<div>").appendTo(this.$el);
		this.list = listEditor(this.$list.get(0));
		this.editor = null;
		this.at = null;
		var detach = _.bind(this.detach, this);
		this.listeners = {
			blur: detach,
			changeSelection: _.bind(this.changeSel, this),
		};
		this.keys = new HashHandler();
		this.keys.bindKeys({
			Up:           _.bind(this.moveto, this, -1),
			Down:         _.bind(this.moveto, this, +1),
			"Tab|Return": _.bind(this.insert, this),
			Esc:          detach,
		});
	},
	attach: function(editor) {
		if (this.editor !== null) this.detach();
		this.editor = editor;
		editor.keyBinding.addKeyboardHandler(this.keys);
		_.each(this.listeners, function(v, k) {
			this.on(k, v);
		}, editor);
	},
	detach: function(e) {
		_.each(this.listeners, function(v, k) {
			this.removeEventListener(k, v);
		}, this.editor);
		this.editor.keyBinding.removeKeyboardHandler(this.keys);
		this.editor = null;
		this.$el.hide();
	},
	changeSel: function(e) {
		var at = this.editor.selection.getCursor();
		if (at.row != this.at.row)
			return this.detach(e);
		if (at.column - this.at.column !== 1)
			return this.detach(e);
		var prefix = this.prefix(at);
		if (!prefix.length)
			return this.detach(e);
		var props = this.proposals.filter(function(prop) {
			var n = prop.get("name");
			return n !== prefix && n.indexOf(prefix) === 0;
		});
		this.at = _.clone(at);
		at.column -= prefix.length;
		this.reposition(at, props.length);
		this.proposals.reset(props);
	},
	prefix: function(at) {
		if (at.column <= 0) return "";
		var c, l = this.editor.session.getLine(at.row);
		for (c = at.column-1; c > 0; c--) {
			if (l[c].match(/\W/)) break;
		}
		return l.slice(c, at.column);
	},
	insert: function() {
		var selected = this.list.selection.lead.row;
		var prop = this.proposals.at(selected);
		// linestart do not select last
		var at = this.editor.selection.getCursor();
		var prefix = this.prefix(at);
		// if not at line start
		var text = prop.get("name");
		if (prefix.length > 0) {
			if (text.indexOf(prefix) !== 0) {
				console.log("proposal does not start with "+ prefix);
				return;
			}
			text = text.slice(prefix.length);
		}
		this.editor.insert(text);
	},
	moveto: function(delta) {
		var last = this.list.session.getLength();
		var sel = this.list.selection;
		sel.clearSelection();
		var row = sel.lead.row+delta;
		if (row == last) {
			sel.moveCursorFileStart();
		} else if (row == -1) {
			sel.moveCursorFileEnd();
		} else {
			sel.moveCursorTo(row);
		}
	},
	render: function() {
		if (!this.proposals.length)
			return this.detach();
		// build lines from proposals
		var max = 0;
		var lines = this.proposals.map(function(e) {
			var n = e.get("name");
			if (n.length > max)
				max = n.length;
			return n +"\t"+ (e.get("type") || e.get("class"));
		});
		this.list.renderer.setPrintMarginColumn(max+1);
		this.list.session.setTabSize(max+2);
		// update list editor
		this.list.setValue(lines.join("\n"), -1);
		this.$el.show();
		this.list.resize();
		return this;
	},
	setData: function(data) {
		var props = _.filter(data.Proposed, function(e) {
			return e.name != "_";
		});
		var at = this.editor.selection.getCursor();
		this.at = _.clone(at);
		var prefix = this.prefix(at);
		at.column -= prefix;
		this.reposition(at, props.length);
		this.proposals.reset(props);
	},
	reposition: function(at, count) {
		at.column = Math.max(at.column,1);
		var pos = getScreenPixels(this.editor, at);
		var lh = this.list.renderer.lineHeight = this.editor.renderer.lineHeight;
		var h = Math.min(6, count) * (lh + 1.6) + 1;
		if (pos.top + h > window.innerHeight) {
			var b = (window.innerHeight-pos.top)&-1;
			this.el.style.top = '';
			this.el.style.bottom = b + 'px';
		} else {
			var t = (pos.top+lh)&-1;
			this.el.style.top = t + 'px';
			this.el.style.bottom = '';
		}
		// clip scrollbar
		this.$list.width(count > 6 ? 300 : 320);
		this.el.style.height = (h&-1)+'px';
		this.el.style.left = pos.left+'px';
	}
});

$('<link>').attr({
	type: "text/css",
	rel:  "stylesheet",
	href: "/static/lib/completion.css",
}).appendTo($("head").first());

var popup = new Popup();
return {
	popup: popup,
	show: function(editor, data) {
		popup.attach(editor);
		popup.setData(data);
	},
};
});

