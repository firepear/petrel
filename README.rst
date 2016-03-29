***********************
petrel
***********************
Automated socket module
#######################

Asock provides a fire-and-forget way to add TCP, TLS, or Unix socket
interfaces to applications written in Go. It handles network I/O and
dispatches requests from clients. All you need to do is watch its
messaging channel for events you'd like to log or act upon.

Asock passes :code:`golint`, :code:`go vet`, and :code:`go test -race`
cleanly. `Test coverage <http://firepear.net/petrel/coverage.html>`_ is
90.7%.

Current version: 0.20.0 (2016-03-09) (`Release notes <https://github.com/firepear/petrel/blob/master/RELEASE_NOTES>`_)

Install with: :code:`go get firepear.net/petrel`

What is it used for?
====================

Asock's original use case was administrative/backend interfaces over
Unix domain sockets — basically networked command lines.

TCP and TLS connections are now supported as well, and input handling
has been made more flexible so that JSON or other structured data can
be fed into programs. So petrel makes it easy to add call/response type
network interfaces to any piece of software. But that said…

You may not wish to use Asock on the internet just yet
======================================================

Despite support for TLS, Asock does not yet support maximum transfer
size limits, or authentication. This makes it vulnerable to DoSing.

How is it used?
===============

See the `full package documentation
<http://godoc.org/firepear.net/petrel>`_ for complete information on
setup options and usage.

See the example `server
<https://github.com/firepear/petrel/blob/master/example/server.go>`_
and `client
<https://github.com/firepear/petrel/blob/master/example/client.go>`_
for a thoroughly documented worked example of a simple Asock
application.


Source and docs
===============

* Install with: :code:`go get firepear.net/petrel`

* `Package documentation <http://godoc.org/firepear.net/petrel>`_

* `Release notes <https://github.com/firepear/petrel/blob/master/RELEASE_NOTES>`_

* `Coverage report <http://firepear.net/petrel/coverage.html>`_

* `Github <https://github.com/firepear/petrel>`_
