package asock // import "firepear.net/asock"

// Copyright (c) 2014,2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"fmt"
	"net"
	"sync"
	"time"
)


// Message levels control which messages will be sent to as.Msgr
const (
	All = iota
	Conn
	Error
	Fatal
	Version = "0.11.0"
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

// Config holds values to be passed to the constuctor.
//
// For Unix sockets, Sockname takes the form "/path/to/socket". For
// TCP socks, it is either an IPv4 or IPv6 address followed by the
// desired port number ("127.0.0.1:9090", "[::1]:9090").
//
// Timeout is the number of seconds the socket will wait before timing
// out due to inactivity. Set it to zero for no timeout. Set to a
// negative value for a connection which closes after handling one
// request — so a timeout of -3 gives a connection which closes after
// one request or a read wait of 3 seconds, whichever happens first.
//
// Msglvl determines which messages will be sent to the socket's
// message channel. Valid values are asock.All, asock.Conn,
// asock.Error, and asock.Fatal.
type Config struct {
	Sockname string
	Timeout  int
	Msglvl   int
}

// Dispatch is the dispatch table which drives asock's behavior. See
// the package Overview for more info on this and DispatchFunc.
type Dispatch map[string]*DispatchFunc

// DispatchFunc instances are the functions called via Dispatch.
//
// Func is the function to be called.
//
// Argmode determines how the bytestream read from the socket will be
// turned into arguments to Func. Valid values are "split" and
// "nosplit".
type DispatchFunc struct {
	Func    func ([][]byte) ([]byte, error)
	Argmode string
}

// Msg is the format which asock uses to communicate informational
// messages and errors to its host program. See the package Overview
// for more info.
type Msg struct {
	Conn uint   // connection id
	Req  uint   // connection request number
	Code int    // numeric status code
	Txt  string // textual description of Msg
	Err  error  // error (if any) passed along from underlying condition
}

// Error implements the error interface, returning a nicely (if
// blandly) formatted string containing all information present in a
// given Msg.
func (m *Msg) Error() string {
	s := fmt.Sprintf("conn %d req %d status %d", m.Conn, m.Req, m.Code)
	if m.Txt != "" {
		s = s + fmt.Sprintf(" (%s)", m.Txt)
	}
	if m.Err != nil {
		s = s + fmt.Sprintf("; err: %s", m.Err)
	}
	return s
}

// NewTCP returns an instance of Asock which uses TCP networking. It
// takes two arguments: an instance of Config and an instance of
// Dispatch
func NewTCP(c Config, d Dispatch) (*Asock, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", c.Sockname)
	l, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return nil, err
	}
	return commonNew(c, d, l), nil
}

// NewUnix returns an instance of Asock which uses Unix domain
// networking. It takes two arguments: an instance of Config and an
// instance of Dispatch.
func NewUnix(c Config, d Dispatch) (*Asock, error) {
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: c.Sockname, Net: "unix"})
	if err != nil {
		return nil, err
	}
	if c.Timeout == -20707 { // triggers the listener to die for failure testing
		c.Timeout = 0
		l.SetDeadline(time.Now().Add(100 * time.Millisecond))
	}
	return commonNew(c, d, l), nil
}

// commonNew does shared setup work for the constructors (mostly so
// that changes to Asock don't have to be mirrored)
func commonNew(c Config, d Dispatch, l net.Listener) *Asock {
	var w sync.WaitGroup
	q := make(chan bool, 1) // master off-switch channel
	m := make(chan *Msg, 32) // error reporting
	a := &Asock{m, q, &w, c.Sockname, l, d, c.Timeout, c.Msglvl}
	a.w.Add(1)
	go a.sockAccept()
	return a
}

// genMsg creates messages and sends them to the Msgr channel.
func (a *Asock) genMsg(conn, req uint, code, ml int, txt string, err error) {
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
