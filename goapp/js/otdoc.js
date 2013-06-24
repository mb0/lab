// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

define(["ot", "ace/range", "ace/document"], function(ot, range, document) {

function utf8OffsetToPos(lines, off, startrow) {
	if (!startrow) {
		startrow = 0;
	}
	var i, line, j, c, lastRow = lines.length;
	for (i=startrow; i<lastRow; i++) {
		line = lines[i];
		for (j=0; off>0 && j<line.length; j++) {
			c = line.charCodeAt(j);
			if (c > 0x10000) {
				off -= 4;
			} else if (c > 0x800) {
				off -= 3;
			} else if (c > 0x80) {
				off -= 2;
			} else {
				off -= 1;
			}
		}
		if (--off < 0 || i == lastRow-1) {
			return {row: i, column: j};
		}
	}
	return {row: i-1, column: j};
}

function posToRestIndex(lines, pos) { // returns {start, last}
	var start = 0, last = 0;
	var i, c, lastRow = lines.length;
	var startRow = Math.min(pos.row, lastRow);
	for (i=0; i<lastRow; i++) {
		c = ot.utf8len(lines[i]);
		last += c;
		if (i < startRow) {
			start += c;
		} else if (i == startRow) {
			start += ot.utf8len(lines[i].slice(0, pos.column));
		}
	}
	return {start:start+startRow, last:last+i-1};
}

function deltaToOps(lines, delta) { // returns ops
	var idxr = posToRestIndex(lines, delta.range.start);
	var ops = [];
	switch (delta.action) {
	case "removeText":
		ops.push(-ot.utf8len(delta.text));
		break;
	case "removeLines":
		var i, n = 0;
		for (i=0; i<delta.lines.length; i++)
			n -= ot.utf8len(delta.lines[i]);
		ops.push(n-i);
		break;
	case "insertText":
		ops.push(delta.text);
		idxr.last -= ot.utf8len(delta.text);
		break;
	case "insertLines":
		var text = delta.lines.concat(['']).join("\n");
		ops.push(text);
		idxr.last -= ot.utf8len(text);
		break;
	default:
		return ops;
	}
	if (idxr.start) {
		ops.unshift(idxr.start);
	}
	if (idxr.last-idxr.start > 0) {
		ops.push(idxr.last-idxr.start);
	}
	return ops;
}

function applyOps(acedoc, ops) {
	var lines = acedoc.$lines || acedoc.getAllLines();
	var count = ot.count(ops);
	var index = 0, pos = {row:0, column: 0}, op;
	var idxr = posToRestIndex(lines, pos);
	if (count.ret+count.del != idxr.last) {
		throw new Error("The base length must be equal to the document length");
	}
	var cache = {row:0, at:0};
	for (var i=0; i < ops.length; i++) {
		op = ops[i];
		if (!op) {
			continue;
		}
		if (typeof op == "string") {
			pos = utf8OffsetToPos(lines, index - cache.at, cache.row);
			cache = {row: pos.row, at: index - pos.column};
			acedoc.insert(pos, op);
			index += ot.utf8len(op);
		} else if (op > 0) {
			index += op;
		} else if (op < 0) {
			var end = utf8OffsetToPos(lines, index-op-cache.at, cache.row);
			pos = utf8OffsetToPos(lines, index-cache.at, cache.row);
			cache = {row: pos.row, at: index-pos.column};
			acedoc.remove(new range.Range(pos.row, pos.column, end.row, end.column));
		}
	}
}

var Doc = function(id, path) {
	this.Id = id;
	this.Path = path;
	this.Rev = -1;
	this.Status = "subscribe";
	this.Ace = new document.Document("");
	this.User = null;
	
	this.wait = null;
	this.buf = null;
	this.merge = false;
};
Doc.prototype = {
	recvOps: function(ops) {
		var res;
		if (this.wait !== null) {
			res = ot.transform(ops, this.wait);
			ops = res[0], this.wait = res[1];
		}
		if (this.buf !== null) {
			res = ot.transform(ops, this.buf);
			ops = res[0], this.buf = res[1];
		}
		this.merge = true;
		applyOps(this.Ace, ops);
		this.merge = false;
		this.Rev++;
		this.Status = "received";
	},
	ackOps: function(ops) {
		if (this.buf !== null) {
			this.wait = this.buf;
			this.buf = null;
			this.Rev++;
			this.Status = "waiting";
			this.Ace._emit("ops", {otdoc: this, ops: this.wait});
		} else if (this.wait !== null) {
			this.wait = null;
			this.Rev++;
			this.Status = "";
		} else {
			throw new Error("no pending operation");
		}
	},
	init: function(rev, user, text) {
		this.Status = "";
		this.Rev = rev;
		this.User = user;
		this.Ace.setValue(text);
		var doc = this;
		this.Ace.on("change", function(e) {
			if (doc.merge === true) {
				return;
			}
			var ops = deltaToOps(doc.Ace.$lines || doc.Ace.getAllLines(), e.data);
			if (!ops) {
				return;
			}
			if (doc.buf !== null) {
				try {
					doc.buf = ot.compose(doc.buf, ops);
				} catch (err) {
					console.log("compose error", err);
				}
			} else if (doc.wait !== null) {
				doc.buf = ops;
			} else {
				doc.wait = ops;
				doc.Status = "waiting";
				doc.Ace._emit("ops", {otdoc: doc, ops: doc.wait});
			}
		});
		this.Ace._emit("init", {otdoc: doc});
	},
};

return {
	Doc: Doc,
	posToRestIndex: posToRestIndex,
};
});