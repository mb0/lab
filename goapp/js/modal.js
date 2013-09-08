// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package modal provides modal container.

define([], function() {

var Modal = function() {
	this.el = document.createElement("div");
	this.el.id = "modal-layer";
	this.el.style.display = "none";
	this.cont = null;
	document.body.appendChild(this.el);
	//this.el.addEventListener("click", this.detach.bind(this));
};
Modal.prototype = {
	attach: function(cont) {
		if (this.cont !== null) this.detach();
		cont.className += " modal-cont";
		this.cont = cont;
		this.el.appendChild(this.cont);
		this.el.style.display = "";
		this.reposition();
	},
	detach: function() {
		if (!this.cont) {
			return;
		}
		this.el.removeChild(this.cont);
		this.cont = null;
		this.el.style.display = "none";
	},
	reposition: function() {
		this.cont.style.left = ((this.el.clientWidth-this.cont.offsetWidth)/2) +"px";
		this.cont.style.top = ((this.el.clientHeight-this.cont.offsetHeight)/2) +"px";
	},
};

var modal = new Modal();
return {
	modal: modal,
	show: function(cont) {
		modal.attach(cont);
	},
};
});