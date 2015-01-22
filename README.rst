*************************************
asock
*************************************
Automated server management interface
=====================================

Asock provides a fire-and-forget way to add a backend (Unix domain
socket) administrative interface to servers written in Go.

It handles network I/O and dispatches requests from clients. All you
need to do is watch its messaging channel for events you'd like to log
or act upon.

What is it used for?
--------------------

Say you've written a piece of server software which allows people to
store and retrieve information about their yo-yo collections. Yo-yo
data insertion and fetching is its public interface, exposed to the
internet.

As the administrator of yoyod, you may wish to instrument it for
monitoring, you may wish for a way to alter its configuration on the
fly, or other similar things. You could add these capabilities to the
public interface, either hoping that other people don't stumble upon
them, or fencing them off with an authentication scheme of some sort.

But you could also add a second interface, which is only exposed to
the machine the server is running on (typically via Unix domain
sockets), and is expressly for administrative purposes.

This way the public interface has no access to administrative
functions, the administrative interface has no access to public
functions, and the internet has no access to the administrative
interface.

Asock makes it easy to provide that private interface.

How is it used?
---------------

::

    func hollaback(s []string) ([]byte, error){
        // a trivial echo server implementation
        return []byte(strings.Join(s, " ")), nil
    }
    
    func set_things_up() {
        // populate a dispatch table
        d := make(asock.Dispatch)
        d["echo"] = hollaback
        
        // instantiate a socket (/tmp/echosock or /var/run/echosock),
        // with no connection timeout, which will generate maximal
        // informational messages
        as, err := asock.New("echosock", d, 0, asock.All)
        
        // if err is nil, the socket is now up and handling requests.
        // if a client connects and sends a message beginning with
        // "echo", the rest of the message will be dispatched to
        // hollaback(). Its return value will then be sent to the client.
        ...    
    }

    // ...then, in an eventloop elsewhere...
    select {
    case msg := <-as.Msgr:
        // Msgr is the message channel from asock. Handle
        // messages and error notifications here.
    case your_other_stuff:
        ...
    }

See the package doc for complete information on setup options and usage.

Source and docs
---------------

* Current version: 0.6.0 (2015-01-17)

* `Release notes <http://firepear.net/asock/RELEASE_NOTES.txt>`_

* `Package documentation <http://firepear.net:6060/pkg/firepear.net/asock/>`_

* `Coverage report <http://firepear.net/asock/coverage.html>`_

* `Issue tracker <https://firepear.atlassian.net/browse/AD>`_
  
* Source repo: :code:`git://firepear.net/asock.git`


If you have questions, suggestions, or problem reports, file a ticket
at the link above or send mail to shawn@firepear.net.
