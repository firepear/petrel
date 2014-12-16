/*
Package adminsock provides a Unix domain socket -- with builtin
request dispatch -- for administration of a daemon.

Dispatch of requests is done by defining a function for each request
you want adminsock to handle. Those functions are then added to an
instance of adminsock.Dispatch, which is passed to the adminsock
constructor. As an example, consider an echo request handler:

    func hollaback(s []string) ([]byte, error){
        return []byte(strings.Join(s, " ")), nil
    }
    
    func main() {
        d := make(adminsock.Dispatch)
        d["echo"] = hollaback
        q, e, err := adminsock.New(d, -1)
        ...
    }

The Dispatch map keys are matched against the first word on each line
of text being read from the socket. Given the above example, if we
sent "echo foo bar baz" to the socket, then hollaback() would be
invoked with:

    []string{"foo", "bar", "baz"}

And it would return:

    []byte("foo bar baz"), nil

All functions added to the Dispatch map must have the signature

    func ([]string) ([]byte, error)

Assuming that the returned error is nil, []byte will be written to the
socket as a response. If error is non-nil, then its stringification
will be sent, preceeded by "ERROR: ".

If the first word of a request does not match the Dispatch, an
unrecognized request error will be sent.

Adminsock's constructor returns two channels and an error. If the
error is not nil, you do not have a working socket.

The first channel is the "quitter" socket. Writing a boolean value to
it will shut down the socket and terminate any long-lived connections.

The second channel is the "error" socket. TODO define error types and
explain them here: fatals, informational, ???  */
package adminsock
