/*
Package petrel provides a simple-to-use networking server and client,
with builtin request dispatch. It is intended to be unobtrusive and
easy to integrate into applications.

QUICK START

Instantiate a handler:

HOW IT WORKS

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

When Server.Quit() is called, the instance stops accepting new
connections, and waits for all existing connections to terminate.

If the Server was configured with long timeouts (or no timeout at
all), then Quit() may block for a long time.

Once Quit() returns, the Server is fully shut down. If you are
recovering from a listener socket error (code 599), it is safe to
spawn a new Server at this point.
*/
package petrel
