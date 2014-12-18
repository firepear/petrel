package adminsock

// Copyright (c) 2014 Shawn Boyette. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// Dispatch is the dispatch table which drives adminsock's behavior.
type Dispatch map[string]func ([]string) ([]byte, error)

// Msg is the format which adminsock uses to communicate informational
// messages and errors to its host program. See the package doc for
// more info.
type Msg struct {
	Txt string
	Err error
}

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
func New(d Dispatch, t int) (chan *Msg, chan bool, *sync.WaitGroup, error) {
	var w sync.WaitGroup
	l, err := net.Listen("unix", buildSockName())
	if err != nil {
		return nil, nil, &w, err
	}
	q := make(chan bool, 1)   // master off-switch channel
	m := make(chan *Msg, 8) // error reporting
	w.Add(1)
	go sockAccept(l, m, q, &w)
	return m, q, &w, err
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
