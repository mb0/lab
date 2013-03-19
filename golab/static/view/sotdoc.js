/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["view/sot", "ace/range"], function(sot, acerange) {

function posToRestIndex(doc, pos) { // returns {start, end, last}
	var lines = doc.$lines || doc.getAllLines();
	var start = 0, last = 0;
	var i = 0, lastRow = lines.length;
	var startRow = Math.min(pos.row, lastRow);
	for (;i < lastRow; i++) {
		last += lines[i].length;
		if (i < startRow) {
			start += lines[i].length;
		}
	}
	return {start:start + pos.column + startRow, last:last+i-1};
}
function joinLines(lines) {
	var res = "";
	for (var i=0; i < lines.length; i++) {
		res += lines[i] + "\n";
	}
	return res;
}
function deltaToOps(doc, delta) { // returns ops
	var idxr = posToRestIndex(doc, delta.range.start);
	var ops = [];
	switch (delta.action) {
	case "removeText":
		ops.push(-delta.text.length);
		break;
	case "removeLines":
		ops.push(-joinLines(delta.lines).length);
		break;
	case "insertText":
		ops.push(delta.text);
		idxr.last -= delta.text.length;
		break;
	case "insertLines":
		var lines = joinLines(delta.lines);
		ops.push(lines);
		idxr.last -= lines.length;
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
function applyOps(doc, ops) { // returns error
	var count = sot.count(ops);
	var lines = doc.getLength();
	var last = doc.positionToIndex({row: lines-1, column: doc.getLine(lines-1).length});
	if (count[0]+count[1] != last) {
		return "The base length must be equal to the document length";
	}
	var index = 0, pos, op;
	doc.sotdoc.merge = true;
	for (var i=0; i < ops.length; i++) {
		if (!(op = ops[i])) continue;
		if (typeof op == "string") {
			pos = doc.indexToPosition(index);
			doc.insert(pos, op)
			index += op.length;
		} else if (op > 0) {
			index += op;
		} else if (op < 0) {
			pos = doc.indexToPosition(index);
			var end = doc.indexToPosition(index-op);
			doc.remove(new acerange.Range(pos.row, pos.column, end.row, end.column));
		}
	}
	doc.sotdoc.merge = false;
	doc.sotdoc.rev++;
	return null;
}

function recvOps(doc, ops) { // returns error
	var sdoc = doc.sotdoc, res;
	if (sdoc.wait != null) {
		res = sot.transform(ops, sdoc.wait);
		if (res[2] != null) {
			return res[2];
		}
		ops = res[0], sdoc.wait = res[1];
	}
	if (sdoc.buf != null) {
		res = sot.transform(ops, sdoc.buf);
		if (res[2] != null) {
			return res[2];
		}
		ops = res[0], sdoc.buf = res[1];
	}
	return applyOps(doc, ops);
}

function emitOps(doc, ops) {
	doc._emit("sotops", {data: {
		Id:  doc.sotdoc.id,
		Rev: doc.sotdoc.rev,
		Ops: doc.sotdoc.wait = ops,
	}});
}

function ackOps(doc, ops) { // returns error
	if (doc.sotdoc.buf != null) {
		doc.sotdoc.rev++;
		emitOps(doc, doc.sotdoc.buf);
		doc.sotdoc.buf = null;
	} else if (doc.sotdoc.wait != null) {
		doc.sotdoc.rev++;
		doc.sotdoc.wait = null;
	} else {
		return "no pending operation";
	}
	return null;
}

function install(id, rev, doc) {
	var sdoc = doc.sotdoc = {
		id: id,
		rev: rev,
		wait: null,
		buf: null,
		merge: false,
	};
	doc.on("change", function(e) {
		if (sdoc.merge === true) {
			return;
		}
		var ops = deltaToOps(doc, e.data);
		if (!ops) return;
		if (sdoc.buf != null) {
			var res = sot.compose(sdoc.buf, ops);
			if (res[1] != null) {
				console.log("compose error", res);
			}
			sdoc.buf = res[0];
		} else if (sdoc.wait != null) {
			sdoc.buf = ops;
		} else {
			emitOps(doc, ops);
		}
	});
	return sdoc;
};

return {
	install: install,
	applyOps: applyOps,
	recvOps: recvOps,
	ackOps: ackOps,
};
});
