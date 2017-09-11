************************
petrel
************************

Petrel is a non-HTTP toolkit for adding network capabilities to
programs. With it you can define APIs/RPCs of arbitrary complexity,
using any data format you like. Here are some key features:

* Works over Unix domain sockets or TCP.

* Optimized for programmer ease-of-use, but has been proven decently
  performant in real-world use.

* Security-conscious design:

  * TLS support for link security and/or client authentication.

  * HMAC support for message verification.

  * Message length limits to protect against memory exhaustion
    denial-of-service attacks.

* Petrel has no external dependencies, and passes :code:`golint`,
  :code:`go vet`, and :code:`go test -race` cleanly.
.. image:: https://goreportcard.com/badge/firepear.net/petrel
  :target: https://goreportcard.com/report/firepear.net/petrel

The current version is 0.30.0 (2016-11-29).

* Install: :code:`go get firepear.net/petrel`

* `Release notes <https://github.com/firepear/petrel/blob/master/RELEASE_NOTES>`_

* `Package documentation <http://godoc.org/firepear.net/petrel>`_

* `Coverage report <http://firepear.net/petrel/coverage.html>`_

* `Github <https://github.com/firepear/petrel>`_

There is a `companion Javascript client library
<https://github.com/firepear/petreljs>`_ under development, but it is
not ready for prime time and requires a websocket-to-petrel bridge.

Examples
========

See the demo `server
<https://github.com/firepear/petrel/blob/master/demo/01-basic/server.go>`_ and
`client
<https://github.com/firepear/petrel/blob/master/demo/01-basic/client.go>`_ for
worked examples of Petrel usage.

To see them in action, in one terminal, do ``go run demo/server.go`` to start the example
server.

In another terminal, try a few runs of the client, like::

  go run demo/client.go date
  go run demo/client.go echo whatever you feel like typing here
  go run demo/client.go
  go run demo/client.go foobar

When you're done, kill the server with C-c in its terminal.
