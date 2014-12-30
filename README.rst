*************************************
adminsock
*************************************
Automated server management interface
=====================================

Adminsock provides a fire-and-forget way to add a backend (Unix domain
socket) administrative interface to servers written in Go.

It handles network I/O and dispatches requests from clients. All you
need to do is watch its messaging channel for events you'd like to log
or act upon.

::

    // a trivial echo server implementation
    func hollaback(s []string) ([]byte, error){
        return []byte(strings.Join(s, " ")), nil
    }
    
    func set_things_up() {
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
        // Msgr is the message channel from adminsock. Handle
        // messages and error notifications here.
    case your_other_stuff:
        ...
    }

See the package doc for complete information on setup options and usage.

* Current version: 0.4.0 (2014-12-29)

* `Package documentation <http://firepear.net:6060/pkg/firepear.net/adminsock/>`_

* `Coverage report <http://firepear.net/adminsock/coverage.html>`_

* `Issue tracker <https://firepear.atlassian.net/browse/AD>`_
  
* Source repo: :code:`git://firepear.net/adminsock.git`


If you have questions, suggestions, or problem reports, file a ticket
at the link above or send mail to shawn@firepear.net.
