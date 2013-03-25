golab
=====
golab is a Go IDE for Linux.

Install
-------
Requires Linux and Go tip.

	go get github.com/mb0/lab/golab
	echo "yay! magic!"

Basic CLI
---------
Flag -work='./...' specifies a path list to your packages.
The default uses the current directory and all it child packages.

	cd $GOROOT
	golab -work=src/pkg/bytes:./src/pkg/hash/...:$GOROOT/src/pkg/log

Features:
 * Automatically installs and tests your packages on change.
 * Prints colored reports to stdout.

Html5 UI
--------
Flag -http starts a http server at localhost:8910.
Flag -addr=:80 uses another server address.

	cd ~/go/src/github.com/mb0/lab
	golab -http

Features:
 * Report view for go errors and test failures with links to sources.
 * Ace editor with gentle highlights and error markers for go, js and css.
 * Document collaboration with operational transformation.
 * External filesystem changes to open documents are merged.
 * godoc  Ctrl+Alt+Click on imports in go source files opens the doc view.
 * gofmt  Ctrl+Shift+F changes the document (does not save to disk)
 * gocode Ctrl+Space shows gocode completion proposals if installed.
 * Unity web launcher integration.

I recommend using the Chrome browser, because the visual feedback seems faster than other browsers.

Feedback
--------
Yes please!
 * https://github.com/mb0/lab/issues
 * http://godoc.org/github.com/mb0/lab

I lost the job and place where i lived just now and need to find new work in Germany fast.
I would even work the first month for food and a place on a sofa. Please send me a mail mb0@mb0.org

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

Recycled code attribution // was easier than adapting to golab
 * ot.js (c) Tim Baumann (MIT License)
