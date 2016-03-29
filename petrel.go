package petrel // import "firepear.net/petrel"

// Copyright (c) 2014-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"crypto/tls"
	"fmt"
	"os"
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
	Pkgname = "petrel"
	Version = "0.21.0"
)

// Handler is a Petrel instance.
type Handler struct {
	// Msgr is the channel which receives notifications from
	// connections.
	Msgr chan *Msg
	q    chan bool
	w    *sync.WaitGroup
	s    string        // socket name
	l    net.Listener  // listener socket
	d    dispatch      // dispatch table
	t    time.Duration // timeout
	ml   int           // message level
}

// AddHandlerFunc adds a handler function to the Handler instance.
//
// argmode has two legal values: "split" and "nosplit"
func (a *Handler) AddHandler(name string, argmode string, df DispatchFunc) error {
	if _, ok := a.d[name]; ok {
		return fmt.Errorf("handler '%v' already exists", name)
	}
	if argmode != "split" && argmode != "nosplit" {
		return fmt.Errorf("invalid argmode '%v'", argmode)
	}
	a.d[name] = &dispatchFunc{df, argmode}
	return nil
}

// genMsg creates messages and sends them to the Msgr channel.
func (a *Handler) genMsg(conn, req uint, code, ml int, txt string, err error) {
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

// Quit handles shutdown and cleanup for petrel instance, including
// waiting for any connections to terminate. When it returns, all
// connections are fully shut down and no more work will be done.
func (a *Handler) Quit() {
	a.q <- true
	a.l.Close()
	a.w.Wait()
	close(a.q)
	close(a.Msgr)
}

// Msg is the format which petrel uses to communicate informational
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

// Config holds values to be passed to the constuctor.
type Config struct {
	// For Unix sockets, Sockname takes the form
	// "/path/to/socket". For TCP socks, it is either an IPv4 or IPv6
	// address followed by the desired port number ("127.0.0.1:9090",
	// "[::1]:9090").
	Sockname string

	// Timeout is the number of milliseconds the socket will wait
	// before timing out due to inactivity. Default (zero) is no
	// timeout.
	Timeout int64

	// Buffer sets how many instances of Msg may be queued in
	// Handler.Msgr. Defaults to 32.
	Buffer int

	// Msglvl determines which messages will be sent to the socket's
	// message channel. Valid values are petrel.All, petrel.Conn,
	// petrel.Error, and petrel.Fatal.
	Msglvl int
}

// dispatch is the dispatch table which drives petrel's behavior. See
// the package Overview for more info on this and DispatchFunc.
type dispatch map[string]*dispatchFunc

// DispatchFunc instances are the functions called via Dispatch.
type DispatchFunc func ([][]byte) ([]byte, error)

// DispatchFunc instances are the functions called via Dispatch.
type dispatchFunc struct {
	// df is the function to be called.
	df DispatchFunc

	// argmode can be "split" or "nosplit". It determines how the
	// bytestream read from the socket will be turned into arguments
	// to Func.
	//
	// Given the input `"echo echo" foo "bar baz" quux`, a function
	// with an Argmode of "nosplit" will receive an arguments list of
	//
	//    []byte{[]byte{`foo "bar baz" quux`}}
	//
	// A fuction with Argmode "split" would get:
	//
	//    []byte{[]byte{`foo`}, []byte{`bar baz`}, []byte{`quux`}}
	argmode string
}

// NewTCP returns a Handler which uses TCP networking.
func NewTCP(c *Config) (*Handler, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", c.Sockname)
	l, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return nil, err
	}
	return commonNew(c, l), nil
}

// NewTLS returns a Handler which uses TCP networking,
// secured with TLS.
func NewTLS(c *Config, t *tls.Config) (*Handler, error) {
	l, err := tls.Listen("tcp", c.Sockname, t)
	if err != nil {
		return nil, err
	}
	return commonNew(c, l), nil
}

// NewUnix returns a Handler which uses Unix domain
// networking. Argument `p` is the Unix permissions to set on the
// socket (e.g. 770)
func NewUnix(c *Config, p uint32) (*Handler, error) {
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: c.Sockname, Net: "unix"})
	if err != nil {
		return nil, err
	}
	err = os.Chmod(c.Sockname, os.FileMode(p))
	if err != nil {
		return nil, err
	}
	return commonNew(c, l), nil
}

// commonNew does shared setup work for the constructors (mostly so
// that changes to Handler don't have to be mirrored)
func commonNew(c *Config, l net.Listener) *Handler {
	// spawn a WaitGroup and add one to it for a.sockAccept()
	var w sync.WaitGroup
	w.Add(1)
	// set c.Buffer to the default if it's zero
	if c.Buffer < 1 {
		c.Buffer = 32
	}
	// create the Handler, start listening, and return
	a := &Handler{make(chan *Msg, c.Buffer),
		make(chan bool, 1),
		&w,
		c.Sockname,
		l, make(dispatch),
		time.Duration(c.Timeout) * time.Millisecond,
		c.Msglvl,
	}
	go a.sockAccept()
	return a
}
