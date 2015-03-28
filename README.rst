***********************
asock
***********************
Automated socket module
#######################

Asock provides a fire-and-forget way to add plain socket interfaces to
servers written in Go. It handles network I/O and dispatches requests
from clients. All you need to do is watch its messaging channel for
events you'd like to log or act upon.

Current version: 0.12.0 (2015-03-28) (`Release notes <https://github.com/firepear/asock/blob/master/RELEASE_NOTES>`_)

What is it used for?
====================

Asock's original use case was administrative/backend interfaces over
Unix domain sockets — basically networked command lines.

Now TCP connections are supported as well, and input handling has been
made more flexible so that JSON or other structured data can be fed
into programs. So asock makes it easy to add call/response type
network interfaces to any piece of software. But that said…

Do not use asock on the internet!
---------------------------------

Asock is not yet secure or hardened. It does not support encryption,
and it has no concept of connection authentication, or rate-limiting,
or maximum transfer sizes. It is not currently safe to use on networks
which are publicly routable!

How is it used?
===============

See the `full package documentation
<http://godoc.org/firepear.net/asock>`_ for complete information on
setup options and usage. This is just a sketch to give some idea of
usage.

::

    func hollaback(args [][]byte) ([]byte, error) {
        return args[0], nil
    }
    
    func set_things_up() {
        // populate a dispatch table
        d := make(asock.Dispatch)
        d["echo"] = &DispatchFunc{hollaback, "nosplit"}
        
        // instantiate a socket (/tmp/echosock or /var/run/echosock),
        // with no connection timeout, which will generate maximal
        // informational messages
        c := Config{"/tmp/echosock.sock", 0, asock.All}
        as, err := asock.NewUnix(c, d)
        
        // if err is nil, the socket is now up and handling requests.
        // if a client connects and sends a message beginning with
        // "echo", the rest of the message will be dispatched to
        // hollaback(). Its return value will then be sent to the client.
        ⋮
    }

    // ...then, in an eventloop elsewhere...
    select {
    case msg := <-as.Msgr:
        // Msgr is the message channel from asock. Handle
        // messages and error notifications here.
    case your_other_stuff:
        ⋮
    }


Source and docs
===============

* Install with: :code:`go get firepear.net/asock`

* `Package documentation <http://godoc.org/firepear.net/asock>`_

* `Release notes <https://github.com/firepear/asock/blob/master/RELEASE_NOTES>`_

* `Coverage report <http://firepear.net/asock/coverage.html>`_

* `Github <https://github.com/firepear/asock>`_
