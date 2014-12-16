package adminsock

// Copyright (c) 2014 Shawn Boyette. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"fmt"
	"net"
	"os"
)

// Dispatch is the dispatch table which drives adminsock's behavior.
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
	var err error
	sockname := fmt.Sprintf("%v-%v.sock", os.Args[0], os.Getpid())
	if os.Getuid() == 0 {
		l, err = net.Listen("unix", "/var/run/" + sockname)
	} else {
		l, err = net.Listen("unix", "/tmp/" + sockname)
	}
	if err != nil {
		return nil, nil, err
	}
	q := make(chan bool, 1)   // master off-switch channel
	e := make(chan error, 32) // error reporting
	go sockAccept(l, q, e)
	return q, e, err
}
