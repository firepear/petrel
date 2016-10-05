package petrel

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

// Server is a Petrel server instance.
type Server struct {
	// Msgr is the channel which receives notifications from
	// connections.
	Msgr chan *Msg
	q    chan bool
	w    *sync.WaitGroup
	s    string        // socket name
	l    net.Listener  // listener socket
	d    dispatch      // dispatch table
	t    time.Duration // timeout
	rl   int32         // request length
	ml   int           // message level
	li   bool          // log ip flag
}

// AddFunc adds a handler function to the Server instance.
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
func (h *Server) AddFunc(name string, mode string, df DispatchFunc) error {
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
func (h *Server) genMsg(conn, req uint, p *Perr, xtra string, err error) {
	// if this message's level is below the instance's level, don't
	// generate the message
	if p.Lvl < h.ml {
		return
	}
	txt := p.Txt
	if xtra != "" {
		txt = fmt.Sprintf("%s: [%s]", txt, xtra)
	}
	h.Msgr <- &Msg{conn, req, p.Code, txt, err}
}

// Quit handles shutdown and cleanup, including waiting for any
// connections to terminate. When it returns, all connections are
// fully shut down and no more work will be done.
func (h *Server) Quit() {
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

// ServerConfig holds values to be passed to server constuctors.
type ServerConfig struct {
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

	// Reqlen is the maximum number of bytes in a single read from the
	// network. If a request exceeds this limit, the connection will
	// be dropped. The default (0) is unlimited.
	Reqlen int

	// Buffer sets how many instances of Msg may be queued in
	// Server.Msgr. If more show up while the buffer is full, they
	// are dropped on the floor to prevent the Server from
	// blocking. Defaults to 32.
	Buffer int

	// Msglvl determines which messages will be sent to the socket's
	// message channel. Valid values are All, Conn,
	// Error, and Fatal.
	Msglvl int

	// LogIP determines if the IP of clients is logged on connect.
	LogIP bool
}

// DispatchFunc is the type which functions passed to Server.AddFunc
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

// NewTCP returns a Server which uses TCP networking.
func TCPServer(c *Config) (*Server, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", c.Sockname)
	l, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return nil, err
	}
	return commonNew(c, l), nil
}

// NewTLS returns a Server which uses TCP networking,
// secured with TLS.
func TLSServer(c *Config, t *tls.Config) (*Server, error) {
	l, err := tls.Listen("tcp", c.Sockname, t)
	if err != nil {
		return nil, err
	}
	return commonNew(c, l), nil
}

// NewUnix returns a Server which uses Unix domain
// networking. Argument `p` is the Unix permissions to set on the
// socket (e.g. 770)
func UnixServer(c *Config, p uint32) (*Server, error) {
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
// that changes to Server don't have to be mirrored)
func commonNew(c *Config, l net.Listener) *Server {
	// spawn a WaitGroup and add one to it for h.sockAccept()
	var w sync.WaitGroup
	w.Add(1)
	// set c.Buffer to the default if it's zero
	if c.Buffer < 1 {
		c.Buffer = 32
	}
	// create the Server, start listening, and return
	h := &Server{make(chan *Msg, c.Buffer),
		make(chan bool, 1),
		&w,
		c.Sockname,
		l, make(dispatch),
		time.Duration(c.Timeout) * time.Millisecond,
		int32(c.Reqlen),
		c.Msglvl,
		c.LogIP,
	}
	go h.sockAccept()
	return h
}
