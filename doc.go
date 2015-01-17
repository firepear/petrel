/*
Package adminsock provides a Unix domain socket -- with builtin
request dispatch -- for administration of a daemon.

COMMAND DISPATCH

Consider this example, showing an instance of adminsock being setup as
an echo server.

    func hollaback(s []string) ([]byte, error){
        // a trivial echo server implementation
        return []byte(strings.Join(s, " ")), nil
    }
    
    func set_things_up() {
        // populate a dispatch table
        d := make(adminsock.Dispatch)
        d["echo"] = hollaback
        
        // instantiate a socket (/tmp/echosock or /var/run/echosock),
        // with no connection timeout, which will generate maximal
        // informational messages
        as, err := adminsock.New("echosock", d, 0, adminsock.All)
        ...
    }

A function is defined for each request which adminsock will handle.
Here there is just the one, hollaback().

These functions are added to an instance of adminsock.Dispatch, which
is passed to adminsock.New(). Functions added to the Dispatch map must
have the signature

    func ([]string) ([]byte, error)

The Dispatch map keys form the command set that the instance of
adminsock understands. They are matched against the first word of text
being read from the socket.

Given the above example, if

    echo foo bar baz

was sent to the socket, then hollaback() would be invoked with:

    []string{"foo", "bar", "baz"}

And it would return:

    []byte("foo bar baz"), nil

If error is nil (as it is here), then the returned byteslice will be
written to the socket as a response.

If error is non-nil, then a message about an internal error having
occurred is sent (no program state is exposed to the client).

If the first word of a request does not match a key in the Dispatch
map, an unrecognized command error will be sent. This message will
contain a list of all known commands. It is left to the user to
provide more comprehensive help.

MONITORING

Servers are typically event-driven and adminsock is designed around
this assumption. Once instantiated, all that needs to be done is
monitoring the Msgr channel. Somewhere in your code, there should be
something like:

    select {
    case msg := <-as.Msgr:
        // Handle adminsock notifications here.
    case your_other_stuff:
        ...
    }

Msgr receives instances of Msg, each of which contains a connection
number, a request number, a status code, a textual description, and an
error.

The connection and request numbers (Msg.Conn, Msg.Req) are included
solely for your client tracking/logging use.

As with HTTP, the status code tells you both generally and
specifically what has occured.

    Code Text                                      Classification
    ---- ----------------------------------------- --------------
     100 client connected                          Informational
     101 dispatching '%v'                                "
     197 ending session                                  " 
     198 client disconnected                             "
     199 terminating listener socket                     "
     200 reply sent                                Success
     400 bad command '%v'                          Client error
     500 request failed                            Server Error
     501 deadline set failed; disconnecting client       "
     599 read from listener socket failed                "

The message level argument to New() controls which messages are sent
to Msgr, but it does not map to a range of codes.

    * Fatal is Adminsock fatal errors only (599)
    * Error adds all other Adminsock errors (all 500s)
    * Conn adds messages about connection opens/closes
    * All adds everything else

Adminsock does not throw away or hide information, so messages which
are not errors according to this table may have a Msg.Err value other
than nil. Client disconnects, for instance, pass along the socket read
error which triggered them. Always test the value of Msg.Err before
using it.

Msgr is a buffered channel, capable of holding 32 Msgs. If Msgr fills
up, new messages will be dropped on the floor to avoid blocking. The
one exception to this is a message with a code of 599, which indicates
that the listener socket itself has stopped working. 

If a message with code 599 is received, immediately halt the adminsock
instance as described in the next section.

SHUTDOWN AND CLEANUP

To halt an adminsock instance, call

    as.Quit()

This will immediately stop the instance from accepting new
connections, and will then wait for all existing connections to
terminate.

Be aware that if the instance was created with very long connection
timeouts (or no timeout at all), then Quit() will block for an
indeterminate length of time.

Once Quit() returns, the instance will have no more execution threads
and will exist only as a reference to an Adminsock struct.

If you are recovering from a listener socket error (a message with
code 599 was received), it is now safe to spawn a new instance if you
wish to do so:

    case msg := <- as.Msgr:
        if msg.Code == 599 {
            as.Quit()
            as = adminsock.New(...)
        }

*/
package adminsock
