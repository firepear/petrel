************************
petrel
************************
Networked dispatch table
########################

Petrel is a non-HTTP toolkit for adding network capabilities to
programs.

Petrel supports Unix domain sockets and TCP connections (with TLS, if
you like).

Petrel has no external dependencies, and passes :code:`golint`,
:code:`go vet`, and :code:`go test -race` cleanly. `Test coverage
<http://firepear.net/petrel/coverage.html>`_ is 90.7%.

Petrel does not handle authentication. That's up to your application.

OLD

Petrel provides a fire-and-forget way to add TCP, TLS, or Unix socket
interfaces to applications written in Go. It handles network I/O and
dispatches requests from clients. All you need to do is watch its
messaging channel for events you'd like to log or act upon.

Current version: 0.22.0 (2016-03-30) (`Release notes <https://github.com/firepear/petrel/blob/master/RELEASE_NOTES>`_)

Install with: :code:`go get firepear.net/petrel`


What is it used for?
====================

Petrel's original use case was administrative/backend interfaces to
daemons, over Unix domain sockets.

TCP and TLS connections are now supported as well, and input handling
has been made more flexible so that JSON or other structured data can
be fed into programs.

So petrel makes it easy to add call/response type network interfaces
to any piece of software.

Despite support for TLS, Petrel does not yet support maximum transfer
size limits. This makes it vulnerable to DoSing, so you may not want
to use it on public networks at this time. (This is coming soon!)

Source and docs
===============

* Install with: :code:`go get firepear.net/petrel`

* `Package documentation <http://godoc.org/firepear.net/petrel>`_

* `Release notes <https://github.com/firepear/petrel/blob/master/RELEASE_NOTES>`_

* `Coverage report <http://firepear.net/petrel/coverage.html>`_

* `Github <https://github.com/firepear/petrel>`_
