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
        m, q, w, err := adminsock.New(d, -1)
        // See 'USE' section for info about New()
        ...
    }

A function is defined for each request which adminsock will
handle. Those functions are added to an instance of
adminsock.Dispatch, which is passed to adminsock.New().

The Dispatch map keys are matched against the first word on each line
of text being read from the socket. Given the above example, if we
sent "echo foo bar baz" to the socket, then hollaback() would be
invoked with:

    []string{"foo", "bar", "baz"}

And it would return:

    []byte("foo bar baz"), nil

The functions added to the Dispatch map must have the signature

    func ([]string) ([]byte, error)

If error is nil, then []byte will be written to the socket as a
response. If error is non-nil, then its stringification will be sent..

If the first word of a request does not match the Dispatch, an
unrecognized request error will be sent.

USE

adminsock.New() returns four values: a message channel, a
sync.WaitGroup instance, a "quitter" channel, and an error.  If err is
not nil, then socket setup has failed and you do not have a working
adminsock instance. 

If you get ENOLISTENER on Msgr, call a.Quit() then instantiate a new
adminsock.

*/
package adminsock
