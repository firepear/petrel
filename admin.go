package adminsock // import "firepear.net/adminsock"

// Copyright (c) 2014,2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"fmt"
	"net"
	"os"
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
	s    string       // socket name
	l    net.Listener // listener socket
	d    Dispatch     // dispatch table
	t    int          // timeout
}

// Quit handles shutdown and cleanup for an adminsock instance,
// including waiting for any connections to terminate. When it
// returns, the Adminsock is fully shut down. See the package Overview
// for more info.
func (a *Adminsock) Quit() {
	a.q <- true
	a.l.Close()
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
	Conn int
	Req  int
	Code int
	Txt  string
	Err  error
}

// New takes three arguments: the socket name, an instance of
// Dispatch, and the connection timeout value, in seconds.
//
// If the process is being run as root, the listener socket will be in
// /var/run; else it will be in /tmp.
//
// If the timeout value is zero, connections will never timeout. If
// the timeout is negative, connections will be "one-shot" -- they
// will perform one read, send one response, and then automatically
// close. One-shot connections still set a timeout value, however
// (e.g. -2 produces a one-shot connection which times out after 2
// seconds.
func New(sn string, d Dispatch, t int) (*Adminsock, error) {
	var w sync.WaitGroup
	if os.Getuid() == 0 {
		sn = fmt.Sprintf("/var/run/%v.sock", sn)
	} else {
		sn = fmt.Sprintf("/tmp/%v.sock", sn)
	}
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: sn, Net: "unix"})
	if err != nil {
		return nil, err
	}
	if t == -20707 { // triggers the listener to die for failure testing
		t = 0
		l.SetDeadline(time.Now().Add(100 * time.Millisecond))
	}
	q := make(chan bool, 1) // master off-switch channel
	m := make(chan *Msg, 32) // error reporting
	a := &Adminsock{m, q, &w, sn, l, d, t}
	a.w.Add(1)
	go a.sockAccept()
	return a, nil
}
