package asock // import "firepear.net/asock"

// Copyright (c) 2014,2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
)


// Message levels control which messages will be sent to as.Msgr
const (
	All = iota
	Conn
	Error
	Fatal
	Version = "0.15.0"
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
	t    int64        // timeout
	ml   int          // message level
}

// Config holds values to be passed to the constuctor.
//
// For Unix sockets, Sockname takes the form "/path/to/socket". For
// TCP socks, it is either an IPv4 or IPv6 address followed by the
// desired port number ("127.0.0.1:9090", "[::1]:9090").
//
// Timeout is the number of milliseconds the socket will wait before
// timing out due to inactivity. Set it to zero for no timeout. Set to
// a negative value for a connection which closes after handling one
// request — so a timeout of -25 gives a connection which closes after
// one request or a read wait of 25 milliseconds, whichever happens
// first.
//
// Buffer is the buffer size (in instances of asock.Msg) for
// Asock.Msgr. Must be at least 1.
//
// Msglvl determines which messages will be sent to the socket's
// message channel. Valid values are asock.All, asock.Conn,
// asock.Error, and asock.Fatal.
//
// TLSConfig is the configuration for TLS connections. Required for
// NewTLS(); can be nil for all other cases.
type Config struct {
	Sockname  string
	Timeout   int64
	Buffer    int
	Msglvl    int
	TLSConfig *tls.Config
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
//
// Given the input `"echo echo" foo "bar baz" quux`, a function with
// an Argmode of "nosplit" will receive an arguments list of
// (byteslices shown as strings for readability):
//
//    [][]byte{`foo "bar baz" quux`}
//
// A fuction with Argmode "split" would get:
//
//    [][]byte{`foo`, `bar baz`, `quux`}
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

// NewTCP returns an instance of Asock which uses TCP networking.
func NewTCP(c Config, d Dispatch) (*Asock, error) {
	if c.Buffer < 1 {
		return nil, fmt.Errorf("asock.Config.Buffer must be > 0, but got: %v", c.Buffer)
	}
	tcpaddr, err := net.ResolveTCPAddr("tcp", c.Sockname)
	l, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return nil, err
	}
	return commonNew(c, d, l), nil
}

// NewTLS returns an instance of Asock which uses TCP networking,
// secured with TLS.
func NewTLS(c Config, d Dispatch) (*Asock, error) {
	if c.Buffer < 1 {
		return nil, fmt.Errorf("asock.Config.Buffer must be > 0, but got: %v", c.Buffer)
	}
	l, err := tls.Listen("tcp", c.Sockname, c.TLSConfig)
	if err != nil {
		return nil, err
	}
	return commonNew(c, d, l), nil
}

// NewUnix returns an instance of Asock which uses Unix domain
// networking.
func NewUnix(c Config, d Dispatch) (*Asock, error) {
	if c.Buffer < 1 {
		return nil, fmt.Errorf("asock.Config.Buffer must be > 0, but got: %v", c.Buffer)
	}
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: c.Sockname, Net: "unix"})
	if err != nil {
		return nil, err
	}
	return commonNew(c, d, l), nil
}

// commonNew does shared setup work for the constructors (mostly so
// that changes to Asock don't have to be mirrored)
func commonNew(c Config, d Dispatch, l net.Listener) *Asock {
	// spawn a WaitGroup and add one to it for a.sockAccept()
	var w sync.WaitGroup
	w.Add(1)
	// create the Asock instance, start listening, and return
	a := &Asock{make(chan *Msg, c.Buffer),	make(chan bool, 1),	&w, c.Sockname, l, d, c.Timeout, c.Msglvl}
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
