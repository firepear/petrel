******************************
adminsock
******************************
Automated management interface
==============================

Adminsock provides a fire-and-forget way to add a backend (Unix
domain) administrative interface to servers written in Go.

It handles network I/O and dispatches requests from clients. All you
need to do is watch its messaging channel for events you'd like to log
or act upon.

::

    // a trivial echo server implementation
    func hollaback(s []string) ([]byte, error){
        return []byte(strings.Join(s, " ")), nil
    }
    
    func main() {
        d := make(adminsock.Dispatch)
        d["echo"] = hollaback
        as, err := adminsock.New(d, 0)

        // so long as err is nil, adminsock is now up and handling
        // requests.

   }

See the package doc for more information on setup options and usage..
    
* Repository: :code:`git://firepear.net/adminsock.git`

* `Coverage report <http://firepear.net/adminsock/coverage.html>`_

Send questions, suggestions, or problem reports to shawn@firepear.net
