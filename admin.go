package asock // import "firepear.net/asock"

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


// Message levels control which messages will be sent to as.Msgr
const (
	All = iota
	Conn
	Error
	Fatal
)

// Asock is a handle on an asock instance. It contains the
// Msgr channel, which is the conduit for notifications from the
// instance.
type Asock struct {
	Msgr chan *Msg
	q    chan bool
	w    *sync.WaitGroup
	s    string       // socket name
	l    net.Listener // listener socket
	d    Dispatch     // dispatch table
	t    int          // timeout
	ml   int          // message level
}

// Dispatch is the dispatch table which drives asock's
// behavior. See the package Overview for more info.
type Dispatch map[string]func ([]string) ([]byte, error)

// Msg is the format which asock uses to communicate informational
// messages and errors to its host program. See the package Overview
// for more info.
type Msg struct {
	Conn int
	Req  int
	Code int
	Txt  string
	Err  error
}

// NewUnix returns an instance of Asock. It takes four arguments: the
// socket name; an instance of Dispatch; the connection timeout value,
// in seconds; and the desired messaging level.
//
// If the timeout value is zero, connections will never timeout. If
// the timeout is negative then connections will perform one read,
// send one response, and then be closed. These "one-shot" connections
// still set a timeout value, (e.g. -2 produces a connection which
// times out after 2 seconds.
//
// Valid message levels are: All, Conn, Error, Fatal
//
// If Asock's process is being run as root, the listener socket
// will be in /var/run; else it will be in /tmp.
func NewUnix(sn string, d Dispatch, t, ml int) (*Asock, error) {
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
	a := &Asock{m, q, &w, sn, l, d, t, ml}
	a.w.Add(1)
	go a.sockAccept()
	return a, nil
}

// genMsg creates messages and sends them to the Msgr channel.
func (a *Asock) genMsg(conn, req, code, ml int, txt string, err error) {
	// if this message's level is below the instance's level, don't
	// generate the message
	if ml < a.ml {
		return
	}
	select {
	case a.Msgr <- &Msg{conn, req, code, txt, err}:
	default:
	}
}

// Quit handles shutdown and cleanup for an asock instance,
// including waiting for any connections to terminate. When it
// returns, the Asock is fully shut down. See the package Overview
// for more info.
func (a *Asock) Quit() {
	a.q <- true
	a.l.Close()
	a.w.Wait()
	close(a.q)
	close(a.Msgr)
}
