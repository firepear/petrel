***********************
asock/client
***********************
Basic client for asock
#######################

The :code:`asock/client` package is a basic, synchronous, asock
client, providing TCP, TLS, and Unix domain socket connections.

How is it used?
===============

See the `full package documentation
<http://godoc.org/firepear.net/asock/client>`_ for complete
information on setup options and usage. This is just a sketch to give
some idea of usage.

::

    c, err := client.NewTCP("127.0.0.1:40404", 0)
    if err != nil {
        // handle client creation failure
    }
    
    response, err := c.Dispatch([]byte(""))
    if err != nil {
        // handle dispatch failures
    }
    // do something with response. when done, do:
    c.Close()


Source and docs
===============

* Install with: :code:`go get firepear.net/asock/client`

* `Package documentation <http://godoc.org/firepear.net/asock/client>`_

* `Coverage report <http://firepear.net/asock/client/coverage.html>`_
