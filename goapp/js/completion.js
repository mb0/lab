/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
// package completion provides completion popup for the ace editor.
define(["require", "ace/editor", "ace/virtual_renderer", "ace/keyboard/hash_handler"],
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

var Popup = function() {
	this.el = document.createElement("div");
	this.el.id = "completion-popup";
	this.el.style.display = "none";
	this.listel = document.createElement("div");
	this.el.appendChild(this.listel);
	this.list = listEditor(this.listel);
	this.editor = null;
	this.at = null;
	document.body.appendChild(this.el);
	this.el.addEventListener("click", this.detach.bind(this));
	this.proposals = [];
	this.keys = new HashHandler();
	this.keys.bindKeys({
		Up:           this.moveto.bind(this, -1),
		Down:         this.moveto.bind(this, +1),
		"Tab|Return": this.insert.bind(this),
		Esc:          this.detach.bind(this),
	});
	this.listeners = {
		blur: this.detach.bind(this),
		changeSelection: this.changeSel.bind(this),
	};
};
Popup.prototype = {
	attach: function(editor) {
		if (this.editor !== null) this.detach();
		this.editor = editor;
		editor.keyBinding.addKeyboardHandler(this.keys);
		for (var name in this.listeners) {
			editor.on(name, this.listeners[name]);
		}
	},
	detach: function() {
		if (!this.editor) {
			return;
		}
		for (var name in this.listeners) {
			this.editor.removeEventListener(name, this.listeners[name]);
		}
		this.editor.keyBinding.removeKeyboardHandler(this.keys);
		this.editor = null;
		this.el.style.display = "none";
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
		var props = [];
		for (var i=0; i < this.proposals.length; i++) {
			var p = this.proposals[i];
			if (p.name != prefix && p.name.indexOf(prefix) === 0) {
				props.push(this.proposals[i]);
			}
		}
		this.at = {row: at.row, column: at.column};
		at.column -= prefix.length;
		this.reposition(at, props.length);
		this.proposals = props;
		this.render();
	},
	prefix: function(at) {
		if (at.column <= 0) return "";
		var c, l = this.editor.session.getLine(at.row);
		for (c = at.column; c > 1; c--) {
			if (l[c-1].match(/\W/)) {
				break;
			}
		}
		return l.slice(c, at.column);
	},
	insert: function() {
		var selected = this.list.selection.lead.row;
		var prop = this.proposals[selected];
		// linestart do not select last
		var at = this.editor.selection.getCursor();
		var prefix = this.prefix(at);
		// if not at line start
		var text = prop.name;
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
		if (this.proposals.length === 0) {
			return this.detach();
		}
		// build lines from proposals
		var max = 0;
		var lines = [];
		for (var i=0; i < this.proposals.length; i++) {
			var p = this.proposals[i];
			if (p.name.length > max) {
				max = p.name.length;
			}
			lines.push(p.name +"\t"+ (p.type || p.class));
		}
		this.list.renderer.setPrintMarginColumn(max+1);
		this.list.session.setTabSize(max+2);
		// update list editor
		this.list.setValue(lines.join("\n"), -1);
		this.el.style.display = "block";
		this.list.resize();
		return this;
	},
	setData: function(data) {
		var props = data.Proposed;
		for (var i=0; i < props.length; i++) {
			if (props[i].name == "_") {
				props.splice(i, 1);
				break;
			}
		}
		var at = this.editor.selection.getCursor();
		this.at = {row: at.row, column: at.column};
		at.column -= this.prefix(at);
		this.reposition(at, props.length);
		this.proposals = props;
		this.render();
	},
	reposition: function(at, count) {
		at.column = Math.max(at.column, 1);
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
		this.listel.style.width = count > 6 ? 300 : 320;
		this.el.style.height = (h&-1)+'px';
		this.el.style.left = pos.left+'px';
	},
};

var popup = new Popup();
return {
	popup: popup,
	show: function(editor, data) {
		popup.attach(editor);
		popup.setData(data);
	},
};
});
