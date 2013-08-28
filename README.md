golab
=====
golab is a Go IDE for Linux.

![golab screenshot][screenshot]

Install
-------
Requires Linux and Go 1.1.

	go get github.com/mb0/lab/golab
	echo 'yay! magic!'

Basic Usage
-----------
`golab` watches all files under your goroot and gopath (`go help gopath`).
It automatically installs and tests a list of packages specified by the `-work` flag and prints colored reports to stdout.

Flag `-work` specifies a path list to the packages you are working on.
Multiple paths can be seperated by a colon `:`.
The default `./...` uses the current directory and all it child packages.

Example:

	cd $GOPATH/src/github.com/mb0
	golab -work=../garyburd/go-websocket/websocket:./lab/...

Html5 UI
--------
`golab -http` starts a web interface for reports and collaborative editing of text files.

Flag `-addr=localhost:8910` specifies the http address.

Example:

	cd $GOPATH/src
	golab -http -addr=:80 -work=github.com/mb0/lab/...

Features:
 * Report view for go errors and test failures with links to sources.
 * Ace editor with gentle highlights and error markers for go, js and css.
 * Document collaboration with operational transformation.
 * External filesystem changes to open documents are merged.
 * gofmt  Ctrl+Shift+F changes the document (does not save to disk)
 * gocode Ctrl+Space shows gocode completion proposals if installed.

I recommend using the Chrome browser, because the visual feedback seems faster than other browsers.

Feedback
--------
Yes please!
 * https://github.com/mb0/lab/issues
 * http://godoc.org/github.com/mb0/lab

License
-------
golab is BSD licensed, Copyright (c) 2013 Martin Schnabel

Server code attributions
 * Go (c) The Go Authors (BSD License)
 * go-websocket (c) Gary Burd (Apache License 2.0)

Client code and asset attributions
 * require.js (c) The Dojo Foundation (BSD/MIT License)
 * json2.js by Douglas Crockford (public domain)
 * Underscore (c) Jeremy Ashkenas (MIT License)
 * Zepto (c) Thomas Fuchs (MIT License)
 * Backbone (c) Jeremy Ashkenas (MIT License)
 * Ace (c) Ajax.org B.V. (BSD License)
 * Font Awesome by Dave Gandy (SIL, MIT and CC BY 3.0 License)
 * Qunit (c) jQuery Foundation and others (MIT License)

Recycled code attribution // was easier than adapting to golab
 * ot.js (c) Tim Baumann (MIT License)

[screenshot]: https://raw.github.com/mb0/lab/master/screenshot.png