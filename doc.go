/*
Package adminsock provides a Unix domain socket -- with builtin
request dispatch -- for administration of a daemon.

SETUP

This example shows the setup of an instance of adminsock as an echo
server.

    func hollaback(s []string) ([]byte, error){
        return []byte(strings.Join(s, " ")), nil
    }
    
    func main() {
        d := make(adminsock.Dispatch)
        d["echo"] = hollaback
        as, err := adminsock.New(d, 0)
        // See 'USE' section for info about New()
        ...
    }

A function is defined for each request which adminsock will
handle. Those functions are added to an instance of
adminsock.Dispatch, which is passed to adminsock.New().

The functions added to the Dispatch map must have the signature

    func ([]string) ([]byte, error)

The Dispatch map keys are matched against the first word on each line
of text being read from the socket. Given the above example, if we
sent "echo foo bar baz" to the socket, then hollaback() would be
invoked with:

    []string{"foo", "bar", "baz"}

And it would return:

    []byte("foo bar baz"), nil

If error is nil, then the returned byteslice will be written to the
socket as a response.

If error is non-nil, then a message about an internal error having
occurred is sent (no program state is exposed to the client).

If the first word of a request does not match the Dispatch, an
unrecognized command error will be sent. This message will contain a
list of all known commands.

USE

Concept is: adminsock absorbs & passes on all errors and shuts self
down when needed. Should never cause failures in your code. At worst,
miss a few connection attempts until you handle ENOLISTENER and spin
up a new instance.

*/
package adminsock
