// Package adminsock provides an automated Unix domain socket for
// local administration of a daemon.
package adminsock

// Copyright (c) 2014 Shawn Boyette. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"net"
	"os"
	"os/user"
)

// Dispatch is the dispatch table which drives adminsock's behavior
type Dispatch map[string]func ([]string) ([]byte, error)

// New returns two channels. The first is a write-only channel which,
// on write, will shut down the socket and all active connections. The
// second is a read-only channel which will deliver errors from the
// socket.
func New(d Dispatch) (chan bool, chan error, error) {
	var l net.Listener
	u, err := user.Current()
	if err != nil {
		return nil, nil, err
	}
	if u.Uid == "0" {
		l, err = net.Listen("unix", "/var/run/" + os.Args[0] + ".sock")
	} else {
		l, err = net.Listen("unix", "/tmp/" + os.Args[0] + ".sock")
	}
	if err != nil {
		return nil, nil, err
	}
	q := make(chan bool, 1)   // master off-switch channel
	e := make(chan error, 32) // error reporting
	go sockAccept(l, q, e)
	return q, e, err
}
