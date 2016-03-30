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

// AddFunc adds a handler function to the Handler instance.
//
// 'name' is the command you wish this function do be the dispatchee
// for.
//
// 'mode' has two legal values: "args" (the portion of the request
// following the command is tokenized in the manner of a Unix shell,
// and those tokens are passed to the function 'df') and "blob" (the
// portion of the request following the command is left as-is, and
// passed as one chunk).
//
// 'df' is the name of the function which will be called on dispatch.
func (h *Handler) AddFunc(name string, mode string, df DispatchFunc) error {
	if _, ok := h.d[name]; ok {
		return fmt.Errorf("handler '%v' already exists", name)
	}
	if mode != "args" && mode != "blob" {
		return fmt.Errorf("invalid mode '%v'", mode)
	}
	h.d[name] = &dispatchFunc{df, mode}
	return nil
}

// genMsg creates messages and sends them to the Msgr channel.
func (h *Handler) genMsg(conn, req uint, code, ml int, txt string, err error) {
	// if this message's level is below the instance's level, don't
	// generate the message
	if ml < h.ml {
		return
	}
	select {
	case h.Msgr <- &Msg{conn, req, code, txt, err}:
	default:
	}
}

// Quit handles shutdown and cleanup, including waiting for any
// connections to terminate. When it returns, all connections are
// fully shut down and no more work will be done.
func (h *Handler) Quit() {
	h.q <- true
	h.l.Close()
	h.w.Wait()
	close(h.q)
	close(h.Msgr)
}

// Msg is the format which Petrel uses to communicate informational
// messages and errors to its host program.
type Msg struct {
	// Conn is the connection ID that the Msg is coming from.
	Conn uint
	// Req is the request number that resulted in the Msg.
	Req uint
	// Code is the numeric status indicator.
	Code int
	// Txt is the content/description.
	Txt string
	// Err is the error (if any) passed upward as part of the Msg.
	Err error
}

// Error implements the error interface for Msg, returning a nicely
// (if blandly) formatted string containing all information present.
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
	// Sockname is the location/IP+port of the socket. For Unix
	// sockets, it takes the form "/path/to/socket". For TCP, it is an
	// IPv4 or IPv6 address followed by the desired port number
	// ("127.0.0.1:9090", "[::1]:9090").
	Sockname string

	// Timeout is the number of milliseconds the socket will wait
	// before timing out due to inactivity. Default (zero) is no
	// timeout -- set this to something nonzero if you don't want
	// requests to be allowed to block forever.
	Timeout int64

	// Buffer sets how many instances of Msg may be queued in
	// Handler.Msgr. Defaults to 32. If more show up while the buffer
	// is full, they are dropped on the floor to prevent the Handler
	// from blocking.
	Buffer int

	// Msglvl determines which messages will be sent to the socket's
	// message channel. Valid values are petrel.All, petrel.Conn,
	// petrel.Error, and petrel.Fatal.
	Msglvl int
}

// DispatchFunc is the type which functions passed to Handler.AddFunc
// must match: taking a slice of slices of bytes as an argument and
// returning a slice of bytes and an error.
type DispatchFunc func ([][]byte) ([]byte, error)

// This is our dispatch table
type dispatch map[string]*dispatchFunc

// And this is how we store DispatchFuncs and their modes, in the
// dispatch table.
type dispatchFunc struct {
	df DispatchFunc
	mode string
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
	// spawn a WaitGroup and add one to it for h.sockAccept()
	var w sync.WaitGroup
	w.Add(1)
	// set c.Buffer to the default if it's zero
	if c.Buffer < 1 {
		c.Buffer = 32
	}
	// create the Handler, start listening, and return
	h := &Handler{make(chan *Msg, c.Buffer),
		make(chan bool, 1),
		&w,
		c.Sockname,
		l, make(dispatch),
		time.Duration(c.Timeout) * time.Millisecond,
		c.Msglvl,
	}
	go h.sockAccept()
	return h
}
