/*
Package petrel provides a TCP or Unix domain socket with builtin
request dispatch.

Petrel is not an HTTP service. It directly manages sockets, so it is
self-contained and compact. It is intended to be unobtrusive and easy
to integrate into applications.

Do not use Petrel on public networks. Petrel does not yet limit
request size, which makes it vulnerable to DoS attacks.

BASIC USAGE

Instantiate a handler:

    pc := &petrel.Config{Sockname: "127.0.0.1:9090", Msglvl: Error}
    ph, err := petrel.Handler(pc)

At this point, if 'err' is nil, then the handler is up and listening
for connections. It can't do anything with them though, because it
doesn't know how to handle any requests. To fix this, add some handler
functions:

    ph.AddFunc("CMD_NAME", ["args"|"blob"], FUNC_NAME)

Now if a request comes in beginning with "CMD_NAME", that request will
be dispatched to function FUNC_NAME. More details on this shortly, but
finishing out the basics of using Petrel, we need to monitor the
messages channel:

    // somewhere in app control flow, do something like this
    select {
    case msg := <-ph.Msgr:
        log.Println(msg)
    case other_stuff:
        ...
    }

Thn, when it's time to shut it all down:

   ph.Quit()

HOW IT WORKS

It's all in Handler.AddFunc. Adding functions to the Handler creates a
set of commands that the Handler knows how to service.

When a request comes over the wire, the first chunk of non-whitespace
characters becomes the key which is used to look up which function
should be called. So if we saw this:

    update {"some": "big wad of JSON"}

Then the Handler would check if AddFunc has been called with a 'name'
parameter of "update". If not, then a generic error response is sent
back over the wire.

If it had, then the function would be called, with everything after
"update" becoming the arguments to that function. Whatever the
function returns (data or error) then gets shipped back over the
network.

HANDLER.MSGR AND MSGS

Msgr is a buffered channel, capable of holding 32 Msgs. If the buffer
fills up, new messages are dropped on the floor to avoid blocking.

The exception to this is a message with a code of 599. It is allowed
to block, since it indicates that the listener socket has stopped
working. If a 599 is received, immediately halt the petrel instance.

Msg.Status tells you what has happened.

Which messages are sent to Msgr is determined by petrel.Config.Msglvl.

    * Fatal is fatal errors only (599)
    * Error adds all other Petrel errors (all 500s)
    * Conn adds messages about connection opens/closes
    * All adds everything else

Messages which are not errors according Petrel may have a Msg.Err
value other than nil. Client disconnects for instance, pass along the
socket read error which triggered them.

SHUTDOWN AND CLEANUP

When Handler.Quit() is called, the instance stops accepting new
connections, and waits for all existing connections to terminate.

If the Handler was configured with long timeouts (or no timeout at
all), then Quit() may block for a long time.

Once Quit() returns, the Handler is fully shut down. If you are
recovering from a listener socket error (code 599), it is safe to
spawn a new Handler at this point.
*/
package petrel
