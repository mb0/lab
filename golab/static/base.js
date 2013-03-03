/*
Copyright 2013 Martin Schnabel. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.
*/
define(["backbone"], function() {

var ModelView = Backbone.View.extend({
	initialize: function() {
		this.listenTo(this.model, "change", this.render);
		this.listenTo(this.model, "remove", this.remove);
	},
	render: function() {
		this.$el.html(this.template(this.model));
		return this;
	}
});

var ListView = Backbone.View.extend({
	tagName: "ul",
	initialize: function() {
		this.listenTo(this.collection, "add", this.addOne);
		this.listenTo(this.collection, "reset", this.render);
	},
	render: function() {
		this.collection.each(this.addOne, this);
		return this;
	},
	addOne: function(model) {
		var v = new this.itemView({model: model, collection: this.collection});
		this.$el.append(v.render().el);
	}
});

return {
	ModelView: ModelView,
	ListView: ListView,
	ListItemView: ModelView.extend({tagName: "li"}),
};

});
