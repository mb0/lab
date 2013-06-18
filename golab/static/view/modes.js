/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["backbone"],
function() {

var Mode = Backbone.Model.extend({
	idAttribute: "name", // mode, regex
});

var Modes = Backbone.Collection.extend({
	model: Mode,
	matchPath: function(path) {
		return this.find(function(mode) {
			return path.match(mode.get("regex"));
		}) || this.get("text");
	},
});

return new Modes(_.map({
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
}, function(val, key) {
	return {
		name: key,
		mode: "ace/mode/" + key,
		regex: new RegExp("^.*\\.(" + val + ")$"),
	};
}));
});
