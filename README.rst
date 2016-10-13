************************
petrel
************************

Petrel is a non-HTTP toolkit for adding network capabilities to
programs. Think of it as websockets without the web, or a networked
dispatch table, or a networked command line.

Petrel works over Unix domain sockets or TCP. It optionally provides
TLS for link security, HMAC for message verification, and message
length limits to protect against memory exhaustion.

Petrel does not handle connection/user authentication, sessions, or
the like. That's up to the application.

Petrel is optimized for programmer ease-of-use but is decently
performant in real-world use.

Petrel has no external dependencies, and passes :code:`golint`,
:code:`go vet`, and :code:`go test -race` cleanly.

The current version is 0.26.0 (2016-10-09).

* Install: :code:`go get firepear.net/petrel`

* `Release notes <https://github.com/firepear/petrel/blob/master/RELEASE_NOTES>`_

* `Package documentation <http://godoc.org/firepear.net/petrel>`_

* `Coverage report <http://firepear.net/petrel/coverage.html>`_

* `Github <https://github.com/firepear/petrel>`_

See the demo `server
<https://github.com/firepear/petrel/blob/master/demo/server.go>`_ and
`client
<https://github.com/firepear/petrel/blob/master/demo/client.go`_ for
worked examples of Petrel usage..
