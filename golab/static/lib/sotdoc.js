/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["lib/sot", "ace/range", "ace/document", "backbone"],
function(sot, range, document) {

function utf8OffsetToPos(lines, off, startrow) {
	if (!startrow) startrow = 0;
	var i, line, j, c, lastRow = lines.length;
	for (i=startrow; i<lastRow; i++) {
		line = lines[i];
		for (j=0; off>0 && j<line.length; j++) {
			c = line.charCodeAt(j);
			if (c > 0x10000) off -= 4;
			else if (c > 0x800) off -= 3;
			else if (c > 0x80) off -= 2;
			else off -= 1;
		}
		if (--off < 0 || i == lastRow-1)
			return {row: i, column: j};
	}
	return {row: i-1, column: j};
}

function posToRestIndex(lines, pos) { // returns {start, end, last}
	var start = 0, last = 0;
	var i, c, lastRow = lines.length;
	var startRow = Math.min(pos.row, lastRow);
	for (i=0; i<lastRow; i++) {
		c = sot.utf8len(lines[i]);
		last += c;
		if (i < startRow) {
			start += c;
		} else if (i == startRow) {
			start += sot.utf8len(lines[i].slice(0, pos.column));
		}
	}
	return {start:start+startRow, last:last+i-1};
}
function joinLines(lines) {
	var res = "";
	for (var i=0; i < lines.length; i++) {
		res += lines[i] + "\n";
	}
	return res;
}

function deltaToOps(lines, delta) { // returns ops
	var idxr = posToRestIndex(lines, delta.range.start);
	var ops = [];
	switch (delta.action) {
	case "removeText":
		ops.push(-sot.utf8len(delta.text));
		break;
	case "removeLines":
		var i, n = 0;
		for (i=0; i<delta.lines.length; i++)
			n -= sot.utf8len(delta.lines[i]);
		ops.push(n-i);
		break;
	case "insertText":
		ops.push(delta.text);
		idxr.last -= sot.utf8len(delta.text);
		break;
	case "insertLines":
		var text = joinLines(delta.lines);
		ops.push(text);
		idxr.last -= sot.utf8len(text);
		break;
	default:
		return [];
	}
	if (idxr.start)
		ops.unshift(idxr.start);
	if (idxr.last-idxr.start > 0)
		ops.push(idxr.last-idxr.start);
	return ops;
}

function applyOps(acedoc, ops) { // returns error
	var lines = acedoc.$lines || acedoc.getAllLines();
	var count = sot.count(ops);
	var index = 0, pos = {row:0, column: 0}, op;
	var idxr = posToRestIndex(lines, pos);
	if (count[0]+count[1] != idxr.last) {
		return "The base length must be equal to the document length";
	}
	var cache = {row:0, at:0};
	for (var i=0; i < ops.length; i++) {
		if (!(op = ops[i])) continue;
		if (typeof op == "string") {
			pos = utf8OffsetToPos(lines, index - cache.at, cache.row);
			cache = {row: pos.row, at: index - pos.column};
			acedoc.insert(pos, op);
			index += sot.utf8len(op);
		} else if (op > 0) {
			index += op;
		} else if (op < 0) {
			var end = utf8OffsetToPos(lines, index-op-cache.at, cache.row);
			pos = utf8OffsetToPos(lines, index-cache.at, cache.row);
			cache = {row: pos.row, at: index-pos.column};
			acedoc.remove(new range.Range(pos.row, pos.column, end.row, end.column));
		}
	}
	return null;
}
var Doc = Backbone.Model.extend({
	idAttribute: "Id", // Path, Rev, User, Status, Ace
	initialize: function(opts) {
		this.wait = null;
		this.buf = null;
		this.merge = false;
	},
	recvOps: function(ops) { // returns error
		var res = null;
		if (this.wait !== null) {
			res = sot.transform(ops, this.wait);
			if (res[2] !== null) {
				return res[2];
			}
			ops = res[0], this.wait = res[1];
		}
		if (this.buf !== null) {
			res = sot.transform(ops, this.buf);
			if (res[2] !== null) {
				return res[2];
			}
			ops = res[0], this.buf = res[1];
		}
		this.merge = true;
		var err = applyOps(this.get("Ace"), ops);
		this.merge = false;
		if (err === null) 
			this.set({Rev: this.get("Rev")+1, Status: "received"});
		return err;
	},
	ackOps: function(ops) { // returns error
		var rev = this.get("Rev")+1;
		if (this.buf !== null) {
			this.wait = this.buf;
			this.buf = null;
			this.set({Rev: rev, Status: "waiting"});
			this.trigger("ops", this, this.wait);
		} else if (this.wait !== null) {
			this.wait = null;
			this.set({Rev: rev, Status: ""});
		} else {
			return "no pending operation";
		}
		return null;
	},
	createAce: function(rev, user, text) {
		var acedoc = new document.Document(text);
		var doc = this;
		acedoc.on("change", function(e) {
			if (doc.merge === true) return;
			var lines = acedoc.$lines || acedoc.getAllLines();
			var ops = deltaToOps(lines, e.data);
			if (ops) doc.onChange(ops);
		});
		this.set({
			Status: "",
			Rev: rev,
			Ace: acedoc,
			User: user,
		});
		return acedoc;
	},
	onChange: function(ops) {
		if (this.buf !== null) {
			var res = sot.compose(this.buf, ops);
			if (res[1] !== null) {
				console.log("compose error", res);
				return;
			}
			this.buf = res[0];
		} else if (this.wait !== null) {
			this.buf = ops;
		} else {
			this.wait = ops;
			this.set({Status: "waiting"});
			this.trigger("ops", this, ops);
		}
	}
});

return {
	Doc: Doc,
	posToRestIndex: posToRestIndex,
};
});

