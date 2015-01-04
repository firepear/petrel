package adminsock // import "firepear.net/adminsock"

// Copyright (c) 2014,2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

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
	l    net.Listener
	d    Dispatch
	t    int
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
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: buildSockName(), Net: "unix"})
	if err != nil {
		return nil, err
	}
	if t == -20707 { // triggers the listener to die for failure testing
		t = 0
		l.SetDeadline(time.Now().Add(100 * time.Millisecond))
	}
	q := make(chan bool, 1) // master off-switch channel
	m := make(chan *Msg, 8) // error reporting
	a := &Adminsock{m, q, &w, l, d, t}
	go a.sockAccept()
	return a, nil
}

func buildSockName() string {
	expath := strings.Split(os.Args[0], "/")
	exname := expath[len(expath) - 1]
	if os.Getuid() == 0 {
		return fmt.Sprintf("/var/run/%v-%v.sock", exname, os.Getpid())
	}
	return fmt.Sprintf("/tmp/%v-%v.sock", exname, os.Getpid())
}
