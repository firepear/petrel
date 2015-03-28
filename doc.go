/*
Package asock provides a TCP or Unix domain socket with builtin
request dispatch.

DO NOT USE ASOCK ON THE INTERNET!

Asock is not yet secure or hardened. It does not support encryption,
and it has no concept of connection authentication, or rate-limiting,
or maximum transfer sizes. It is not currently safe to use on networks
which are publicly routable!

COMMAND DISPATCH

Consider this example, showing an instance of asock being setup as
an echo server.

    func hollaback(args [][]byte) ([]byte, error) {
        var hb []byte
        for i, arg := range args {
            hb = append(hb, arg...)
            if i != len(args) - 1 {
                hb = append(hb, byte(32))
            }
        }
        return hb, nil
    }
    
    func set_things_up() {
        // populate a dispatch table
        d := make(asock.Dispatch)
        d["echo"] = &DispatchFunc{hollaback, "split"}
        
        // instantiate a socket with no connection timeout,
        // which will generate maximal informational messages
        c := Config{"/tmp/echosock.sock", 0, asock.All}
        as, err := asock.NewUnix(c, d)
        ...
    }

A function is defined for each request which asock will handle.  (Here
there is just the one, hollaback().) Any such function must have the
signature:

    func ([][]byte) ([]byte, error)

These functions are wrapped in DispatchFunc structs and added to an
instance of Dispatch, which is then passed to the constructor.

The keys of the Dispatch map form the command set that the instance of
asock understands. The first word of each request read from the socket
is treated as the command for that request.

If the input from the socket was:

    echo foo bar baz

then "echo" would be the Dispatch entry to be called, and hollaback()
would be invoked with (showing byteslices as type conversions of
strings for readability):

    []byte{[]byte("foo"), []byte("bar"), []byte("baz")}

And it would return:

    []byte("foo bar baz"), nil

If the error is nil (as it is here), then the returned byteslice will
be written to the socket as a response.

If the error is non-nil, then a message about an internal error having
occurred is sent (no program state is exposed to the client).

If the first word of a request does not match a key in the Dispatch
map, an unrecognized command error will be sent. This message will
contain a list of all known commands. It is left to the user to
provide more comprehensive help.

JSON AND OTHER MONOLITHIC DATA

If a function need to be passed JSON -- or other data which should not
be modified outside your control -- then set Argmode to "nosplit" in
that DispatchFunc.

This will cause the dispatch command to be split off from the data
read over the socket, and the remainder of the data to be passed to
your function as a single byteslice.

Returning to the echo server example, a Dispatch entry for "echo"
using nosplit would be:

    d["echo"] = &DispatchFunc{hollaback, "split"}

With an input of "echo foo bar baz", hollaback() would be called with
the following argument (again, shown with type conversions for
readability):

    []byte{[]byte("foo bar baz")}

DISPATCH EXECUTION

Each connection is handled by its own goroutine, so the overall
operation of asock is asynchronous. This means that Dispatch functions
need to be written in a thread-safe manner.

The connection handler routine itself, however, is synchronous in
operation, so there are no extra complexities or hidden
"gotchas". Asock is also tested with the Go race detector, and there
are no known race conditions within it.

MONITORING

Servers are typically event-driven and asock is designed around
this assumption. Once instantiated, all that needs to be done is
monitoring the Msgr channel. Somewhere in your code, there should be
something like:

    select {
    case msg := <-as.Msgr:
        // Handle asock notifications here.
    case your_other_stuff:
        ...
    }

Msgr receives instances of Msg, each of which contains a connection
number, a request number, a status code, a textual description, and an
error.

Msg implements the error interface, so instances of it will be
automatically stringified when passed to standard printing and logging
functions.

As with HTTP, the status code tells you both generally and
specifically what has occured.

    Code Text                                      Type
    ---- ----------------------------------------- -------------
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

    * Fatal is Asock fatal errors only (599)
    * Error adds all other Asock errors (all 500s)
    * Conn adds messages about connection opens/closes
    * All adds everything else

Asock does not throw away or hide information, so messages which are
not errors according to this table may have a Msg.Err value other than
nil. Client disconnects, for instance, are not treated as an error
condition within asock, but do pass along the socket read error which
triggered them. Always test the value of Msg.Err before using it.

Msgr is a buffered channel, capable of holding 32 Msgs. In general, it
is advised to keep Msgr drained. If Msgr fills up, new messages will
be dropped on the floor to avoid blocking.

The one exception to this is a message with a code of 599, which
indicates that the listener socket itself has stopped working. If a
message with code 599 is received, immediately halt the asock instance
as described in the next section.

SHUTDOWN AND CLEANUP

To halt an asock instance, call

    as.Quit()

This will immediately stop the instance from accepting new
connections, and will then wait for all existing connections to
terminate.

Be aware that if the instance was created with very long connection
timeouts (or no timeout at all), then Quit() will block for an
indeterminate length of time.

Once Quit() returns, the instance will have no more execution threads
and will exist only as a reference to an Asock struct.

If you are recovering from a listener socket error (a message with
code 599 was received), it is now safe to spawn a new instance if you
wish to do so:

    case msg := <- as.Msgr:
        if msg.Code == 599 {
            as.Quit()
            as = asock.New(...)
        }

*/
package asock
