*************************************
adminsock
*************************************
Automated server management interface
=====================================

Adminsock provides a fire-and-forget way to add a backend (Unix
domain) administrative interface to servers written in Go.

It handles network I/O and dispatches requests from clients. All you
need to do is watch its messaging channel for events you'd like to log
or act upon.

::

    // a trivial echo server implementation
    func hollaback(s []string) ([]byte, error){
        return []byte(strings.Join(s, " ")), nil
    }
    
    func somewhere_initty() {
        d := make(adminsock.Dispatch)
        d["echo"] = hollaback
        as, err := adminsock.New(d, 0)
        // if err is nil, adminsock is now up and handling
        // requests. if a client connects and sends a message beginning
        // with "echo", the rest of the message will be dispatched to
        // hollaback(). Its return value will then be sent to the client.
        ...    
    }

    // ...then, in an eventloop elsewhere...
    select {
        case msg := <-as.Msgr:
            // Msgr is the message channel from adminsock. Operational
            // messages and error notifications appear here. Do with them
            // what you will.
        case your_other_stuff:
            ...
        }
    }

See the package doc for complete information on setup options and usage..
    
* `Package documentation <http://firepear.net:6060/pkg/firepear.net/adminsock/>`_

* `Coverage report <http://firepear.net/adminsock/coverage.html>`_

* `Issue tracker <https://firepear.atlassian.net/browse/AD>`_
  
* Repository: :code:`git://firepear.net/adminsock.git`


Send questions, suggestions, or problem reports to shawn@firepear.net
