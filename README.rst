************************
petrel
************************

Petrel is a non-HTTP toolkit for adding network capabilities to
programs. Think of it as websockets without the web, or a networked
dispatch table, or a networked command line.

Petrel is optimized for programmer ease-of-use but has been shown to
be decently performant in real-world use.

Petrel works over Unix domain sockets or TCP. It optionally provides
TLS for link security, HMAC for message verification, and message
length limits to protect against memory exhaustion.

Petrel does not handle connection or user authentication. That's up to
the application.

Petrel has no external dependencies, and passes :code:`golint`,
:code:`go vet`, and :code:`go test -race` cleanly.

The current version is 0.25.0 (2016-10-07). Here are the `release
notes
<https://github.com/firepear/petrel/blob/master/RELEASE_NOTES>`_.

* Install: :code:`go get firepear.net/petrel`

* `Release notes <https://github.com/firepear/petrel/blob/master/RELEASE_NOTES>`_

* `Package documentation <http://godoc.org/firepear.net/petrel>`_

* `Coverage report <http://firepear.net/petrel/coverage.html>`_

* `Github <https://github.com/firepear/petrel>`_
