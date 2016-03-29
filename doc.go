/*
Package petrel provides a TCP or Unix domain socket with builtin
request dispatch.

DO NOT USE PETREL ON PUBLIC NETWORKS YET

Petrel now supports TLS, but it does not yet support connection
authentication, or maximum transfer sizes.

COMMAND DISPATCH

Consider this example, showing an instance of petrel being setup as
an echo server.

    func hollaback(args [][]byte) ([]byte, error) {
        return args[0], nil
    }
    
    func set_things_up() {
        // instantiate a socket with no connection timeout,
        // which will generate maximal informational messages
        c := &Config{
            Sockname: "/tmp/echosock.sock",
            Msglvl: petrel.All,
        }
        as, err := petrel.NewUnix(c, d)
        // now add a handler and we're ready to serve
        err = as.AddHandlerFunc("echo", "nosplit", hollaback)
        ...
    }

A function is defined for each command which this petrel instance will
handle -- here there is just hollaback(), which is assigned to be the
"echo" handler.

All handler functions must be of type

    func ([][]byte) ([]byte, error)

The names given as the first argument to AddHandlerFunc() forms the
command set that the instance of petrel understands. The first word (or
quoted string) of each request read from the socket is treated as the
command for that request.

HANDLERS AND ARGMODES

If we connected to the above server and sent:

    echo foo 'bar baz' quux

then "echo" would be the handler invoked, and hollaback() would be
called with these arguments (showing byteslices as type conversions of
strings for readability):

    []byte{[]byte("foo 'bar baz' quux")}

If, however, we had called AddHandlerFunc() with "split" as its second
argument, then the input following the command ("echo" in this case)
would be split into chunks by word or quoted string. Then, hollaback()
would be called with:

    []byte{[]byte("foo"), []byte("bar baz"), []byte("quux")}

However, hollaback(), as written, would not provide correct output if
we declared its argmode to be "split".

The purpose of argmodes is to support various usecases. If you're
implementing something that behaves like the shell, using "split" will
save you some work. If, however, you're feeding JSON or other data
which should not be cooked by Petrel, then "nosplit" will pass it to
your handler untouched.

HANDLER RETURNS AND ERRORS

If the error returned by the handler is nil (as it is here), then the
returned byteslice will be written to the socket as a response.

If the error is non-nil, then a generic message about an internal
error having occurred is sent. No program state is exposed to the
client, but you would have diagnostic info available to you on the
Msgr channel of your Handler (more about that later).

If the first word of a request does not match the name of a defined
handler, then an unrecognized command error will be sent. This message
will contain a list of all known commands.

HANDLER EXECUTION

Each connection to an instance of petrel is handled by its own
goroutine, so the overall operation of petrel is asynchronous. This
means that handler functions should to be written in a thread-safe
manner.

The management of individual connections, however, is
synchronous. Connections block while waiting on handler functions to
complete.

If you don't want handlers to potentially block forever, set a Timeout
value in the petrel.Config instance that you pass to the constructor.

Petrel is also tested with the Go race detector, and there are no known
race conditions within it.

MONITORING

Servers are typically event-driven and Petrel is designed around
this assumption. Once instantiated, all that needs to be done is
monitoring the Msgr channel. Somewhere in your code, there should be
something like:

    select {
    case msg := <-as.Msgr:
        // Handle petrel notifications here.
    case your_other_stuff:
        ...
    }

Msgr receives instances of Msg.

Msg implements the standard error interface, so instances of it will
be automatically (if generically) stringified when passed to standard
printing and logging functions. The example server demonstrates this.

MSGS

The status code of a Msg tells you what has occured.

    Code Text                                      Type
    ---- ----------------------------------------- -------------
     100 client connected                          Informational
     101 dispatching '%v'                                "
     196 network error                                   "
     197 ending session                                  "
     198 client disconnected                             "
     199 terminating listener socket                     "
     200 reply sent                                Success
     400 bad command '%v'                          Client error
     401 nil request                                     "
     500 request failed                            Server Error
     501 internal error                                  "
     599 read from listener socket failed                "

petrel.Config.Msglvl controls which messages are sent to petrel.Msgr:

    * Fatal is Petrel fatal errors only (599)
    * Error adds all other Petrel errors (all 500s)
    * Conn adds messages about connection opens/closes
    * All adds everything else

Petrel does not throw away or hide information, so messages which are
not errors according to this table may have a Msg.Err value other than
nil. Client disconnects, for instance, are not treated as an error
condition within petrel, but do pass along the socket read error which
triggered them. Always test the value of Msg.Err before using it.

Msgr is a buffered channel, capable of holding 32 Msgs. If the buffer
fills up, new messages are dropped on the floor to avoid blocking.

The one exception to this is a message with a code of 599, which is
allowed to block, since it indicates that the listener socket itself
has stopped working. If a 599 is received, immediately halt the petrel
instance as described in the next section.

SHUTDOWN AND CLEANUP

To halt petrel instance, call

    as.Quit()

This will stop the instance from accepting new connections, and will
then wait for all existing connections to terminate.

If the instance was created with very long connection timeouts (or no
timeout at all), then Quit() will block for an indeterminate length of
time.

Once Quit() returns, the instance will have no more execution threads
and will exist only as a reference to Handler struct.

If you are recovering from a listener socket error (a message with
code 599 was received), it is now safe to spawn a new instance if you
wish to do so:

    case msg := <- as.Msgr:
        if msg.Code == 599 {
            as.Quit()
            as = petrel.New(...)
        }

*/
package petrel
