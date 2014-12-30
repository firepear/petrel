package adminsock

// Copyright (c) 2014 Shawn Boyette <shawn@firepear.net>. All rights
// reserved.  Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Adminsock is a handle on an adminsock instance. It contains the
// Msgr channel, which is the conduit for notifications from the
// instance.
type Adminsock struct {
	Msgr chan *Msg
	q    chan bool
	w    *sync.WaitGroup
}

// Quit handles shutdown and cleanup for an adminsock instance,
// including waiting for any connections to terminate. When it
// returns, the Adminsock is fully shut down. See the package Overview
// for more info.
func (a *Adminsock) Quit() {
	a.q <- true
	a.w.Wait()
	close(a.q)
	close(a.Msgr)
}

// Dispatch is the dispatch table which drives adminsock's
// behavior. See the package Overview for more info.
type Dispatch map[string]func ([]string) ([]byte, error)

// Msg is the format which adminsock uses to communicate informational
// messages and errors to its host program. See the package Overview
// for more info.
type Msg struct {
	Txt string
	Err error
}

// New takes two arguments: an instance of Dispatch, and the
// connection timeout value, in seconds.
//
// If the timeout value is zero, connections will never timeout. If
// the timeout is negative, connections will perform one read, send
// one response, and then automatically close.
//
// The listener socket will be called PROCESSNAME-PID.sock. If the
// process is being run as root, it will be in /var/run; else it will
// be in /tmp.
func New(d Dispatch, t int) (*Adminsock, error) {
	var w sync.WaitGroup
	l, err := net.ListenUnix("unix", &net.UnixAddr{buildSockName(), "unix"})
	if err != nil {
		return nil, err
	}
	if t == -42 { // triggers the listener to die for failure testing
		l.SetDeadline(time.Now().Add(100 * time.Millisecond))
	}
	q := make(chan bool, 1) // master off-switch channel
	m := make(chan *Msg, 8) // error reporting
	w.Add(1)
	go sockAccept(l, d, t, m, q, &w)
	return &Adminsock{m, q, &w}, err
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
