/*
Package adminsock provides a Unix domain socket -- with builtin
request dispatch -- for administration of a daemon.

COMMAND DISPATCH

Consider this example, showing an instance of adminsock being setup as
an echo server.

    func hollaback(s []string) ([]byte, error){
        return []byte(strings.Join(s, " ")), nil
    }
    
    func main() {
        d := make(adminsock.Dispatch)
        d["echo"] = hollaback
        as, err := adminsock.New(d, 0)
        // if err != nil, adminsock is up and listening
        ...
    }

A function is defined for each request which adminsock will handle
(here there is just the one, hollaback()).

These functions are added to an instance of adminsock.Dispatch, which
is passed to adminsock.New(). Functions added to the Dispatch map must
have the signature

    func ([]string) ([]byte, error)

The Dispatch map keys form the command set that the instance of
adminsock understands. They are matched against the first word of text
being read from the socket.

Given the above example, if "echo foo bar baz" was sent to the socket,
then hollaback() would be invoked with:

    []string{"foo", "bar", "baz"}

And it would return:

    []byte("foo bar baz"), nil

If error is nil, then the returned byteslice will be written to the
socket as a response. If error is non-nil, then a message about an
internal error having occurred is sent (no program state is exposed to
the client).

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

Msgr sends instances of Msg, each of which contains string (Msg.Txt)
and an error (Msg.Err). If Msg.Err is nil, then the message is purely
informational (client connects, unknown commands, etc.).

If Msg.Err is not nil, then the message is an actual error being
passed along. Most errors will also be inoccuous (clients dropping
connections, etc.), and adminsock should continue operating with no
intervention.

However, if Msg.Err is not nil and Msg.Txt is "ENOLISTENER", then a
local networking error has occurred and the adminsock's listener
socket has gone away. If this happens, your adminsock instance is no
longer viable; clean it up and spawn a new one. You should, at worst,
drop a few connection attempts.

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

If you are recovering from a ENOLISTENER condition, it's safe at this
point to spawn a new instance:

    case msg := <- as.Msgr:
        if msg.Err != nil && msg.Txt == "ENOLISTENER" {
            as.Quit()
            as = adminsock.New(d, 0)
        }

*/
package adminsock
