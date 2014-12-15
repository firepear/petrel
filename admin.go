// Package adminsock provides a Unix domain socket -- with builtin
// command dispatch -- for administration of a daemon.
//
// In addition to constructing and managing a socket, adminsock
// handles dispatch of requests which arrive on that socket. This is
// done by defining a function for each type of request you want
// adminsock to handle, then adding those functions to an instance of
// adminsock.Dispatch, which is passed to the adminsock constructor.
//
// As an example, consider an echo request handler:
//
//    func hollaback(s []string) ([]byte, error){
//        return []byte(strings.Join(s, " ")), nil
//    }
//
//    func main() {
//        d := make(adminsock.Dispatch)
//        d["echo"] = hollaback
//        q, e, err := adminsock.New(d, -1)
//        ...
//    }
//
// The Dispatch map keys are matched against the first word on each
// line of text being read from the socket. Given the above example,
// if we sent "echo foo bar baz" to the socket, then hollaback() would
// be invoked with an argument of
//
//    []string("foo", "bar", "baz")
//
// and it would return
//
//    []byte("foo bar baz"), nil
//
// All functions added to the Dispatch map must have the signature
//
//    func ([]string) ([]byte, error)
//
// Assuming that the returned error is nil, the []byte will be written
// to the socket as a response. If error is non-nil, then its
// stringification will be sent, preceeded by "ERROR: ".
//
// If the first word of a request does not match the Dispatch, an
// unrecognized request error will be sent.
//
// Adminsock's constructor returns two channels and an error. If the
// error is not nil, you do not have a working socket.
//
// The first channel is the "quitter" socket. Writing a boolean value
// to it will shut down the socket and terminate any long-lived
// connections.
//
// The second channel is the "error" socket. TODO define error types
// and explain them here: fatals, informational, ???
package adminsock

// Copyright (c) 2014 Shawn Boyette. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"fmt"
	"net"
	"os"
	"os/user"
)

// Dispatch is the dispatch table which drives adminsock's behavior
type Dispatch map[string]func ([]string) ([]byte, error)

// New takes two arguments. The first is a Dispatch, which is used to
// drive the behavior of the socket (as described earlier). The second
// is the connection timeout value, in seconds. If this value is zero,
// connections will never timeout. If it is -1, connections will
// accept one line of input, send one response, and close.
//
// New returns two channels. The first is a write-only channel which,
// on write, will shut down the socket and any active connections. The
// second is a read-only channel which will deliver errors from the
// socket.
func New(d Dispatch, t int) (chan bool, chan error, error) {
	var l net.Listener
	u, err := user.Current()
	if err != nil {
		return nil, nil, fmt.Errorf("Could not determine user; cannot create socket: %v", err)
	}
	namepid := fmt.Sprintf("%v-%v", os.Args[0], os.Getpid())
	if u.Uid == "0" {
		l, err = net.Listen("unix", "/var/run/" + namepid + ".sock")
	} else {
		l, err = net.Listen("unix", "/tmp/" + namepid + ".sock")
	}
	if err != nil {
		return nil, nil, err
	}
	q := make(chan bool, 1)   // master off-switch channel
	e := make(chan error, 32) // error reporting
	go sockAccept(l, q, e)
	return q, e, err
}
