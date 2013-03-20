/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["view/sot", "ace/range"], function(sot, acerange) {

function utf8OffsetToPos(doc, off, startrow) {
	var line, lines = doc.$lines || doc.getAllLines();
	if (!startrow) startrow = 0;
	var i, j, c, lastRow = lines.length;
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

function posToRestIndex(doc, pos) { // returns {start, end, last}
	var lines = doc.$lines || doc.getAllLines();
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
function deltaToOps(doc, delta) { // returns ops
	var idxr = posToRestIndex(doc, delta.range.start);
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
function applyOps(doc, ops) { // returns error
	var count = sot.count(ops);
	var index = 0, pos = {row: 0, column: 0}, op;
	var idxr = posToRestIndex(doc, pos);
	if (count[0]+count[1] != idxr.last) {
		return "The base length must be equal to the document length";
	}
	doc.sotdoc.merge = true;
	for (var i=0; i < ops.length; i++) {
		if (!(op = ops[i])) continue;
		if (typeof op == "string") {
			pos = utf8OffsetToPos(doc, index);
			console.log(op, index, pos);
			doc.insert(pos, op);
			index += sot.utf8len(op);
		} else if (op > 0) {
			console.log(op, index);
			index += op;
		} else if (op < 0) {
			pos = utf8OffsetToPos(doc, index);
			var end = utf8OffsetToPos(doc, index-op);
			console.log(op, index, pos, end);
			doc.remove(new acerange.Range(pos.row, pos.column, end.row, end.column));
		}
	}
	doc.sotdoc.merge = false;
	doc.sotdoc.rev++;
	return null;
}

function recvOps(doc, ops) { // returns error
	var sdoc = doc.sotdoc, res;
	if (sdoc.wait !== null) {
		res = sot.transform(ops, sdoc.wait);
		if (res[2] !== null) {
			return res[2];
		}
		ops = res[0], sdoc.wait = res[1];
	}
	if (sdoc.buf !== null) {
		res = sot.transform(ops, sdoc.buf);
		if (res[2] !== null) {
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
	if (doc.sotdoc.buf !== null) {
		doc.sotdoc.rev++;
		emitOps(doc, doc.sotdoc.buf);
		doc.sotdoc.buf = null;
	} else if (doc.sotdoc.wait !== null) {
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
		if (sdoc.buf !== null) {
			var res = sot.compose(sdoc.buf, ops);
			if (res[1] !== null) {
				console.log("compose error", res);
			}
			sdoc.buf = res[0];
		} else if (sdoc.wait !== null) {
			sdoc.buf = ops;
		} else {
			emitOps(doc, ops);
		}
	});
	return sdoc;
}

return {
	install: install,
	applyOps: applyOps,
	recvOps: recvOps,
	ackOps: ackOps,
};
});
