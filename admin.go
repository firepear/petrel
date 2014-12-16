package adminsock

// Copyright (c) 2014 Shawn Boyette. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"fmt"
	"net"
	"os"
	"strings"
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
	l, err := net.Listen("unix", buildSockName())
	if err != nil {
		return nil, nil, err
	}
	q := make(chan bool, 1)   // master off-switch channel
	e := make(chan error, 32) // error reporting
	go sockAccept(l, q, e)
	return q, e, err
}

func buildSockName() string {
	expath := strings.Split(os.Args[0], "/")
	exname := expath[len(expath) - 1]
	if os.Getuid() == 0 {
		return fmt.Sprintf("/var/run/%v-%v.sock", exname, os.Getpid())
	} else {
		return fmt.Sprintf("/tmp/%v-%v.sock", exname, os.Getpid())
	}
}
