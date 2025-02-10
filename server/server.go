package server

// Copyright (c) 2014-2025 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	p "github.com/firepear/petrel"
)

// Message levels control which messages will be sent to h.Msgr
const (
	Debug = iota
	Info
	Error
	Fatal
)

// Server is a Petrel server instance.
type Server struct {
	// Msgr is the channel which receives notifications from
	// connections.
	Msgr chan *Msg
	// Sig is p.Sigchan, made available so apps can waatch it
	Sig chan os.Signal
	q   chan bool     // quit signal socket
	s   string        // socket name
	l   net.Listener  // listener socket
	d   dispatch      // dispatch table
	t   time.Duration // timeout
	rl  uint32        // request length
	ml  int           // message level
	li  bool          // log ip flag
	hk  []byte        // HMAC key
	w   *sync.WaitGroup
}

// Config holds values to be passed to server constuctors.
type Config struct {
	// Sockname is the IP+port of the socket, e.g."127.0.0.1:9090"
	// or "[::1]:9090".
	Sockname string

	// Timeout is the number of milliseconds the Server will wait
	// when performing network ops before timing out. Default
	// (zero) is no timeout. Each connection to the server is
	// handled in a separate goroutine, however, so one blocked
	// connection does not affect any others (unless you run out of
	// file descriptors for new conns).
	Timeout int64

	// Xferlim is the maximum number of bytes in a single read from
	// the network. If a request exceeds this limit, the
	// connection will be dropped. Use this to prevent memory
	// exhaustion by arbitrarily long network reads. The default
	// (0) is unlimited.
	Xferlim uint32

	// Buffer sets how many instances of Msg may be queued in
	// Server.Msgr. Non-Fatal Msgs which arrive while the buffer
	// is full are dropped on the floor to prevent the Server from
	// blocking. Defaults to 32.
	Buffer int

	// Msglvl determines which messages will be sent to the
	// Server's message channel. Valid values: debug, conn, error,
	// fatal.
	Msglvl string

	// LogIP determines if the IP of clients is logged on
	// connect. Enabling IP logging creates a bit of overhead on
	// each connect. If this isn't needed, or if the client can be
	// identified at the application layer, leaving this off will
	// somewhat improve performance in high-usage scenarios.
	LogIP bool

	// HMACKey is the secret key used to generate MACs for signing
	// and verifying messages. Default (nil) means MACs will not
	// be generated for messages sent, or expected for messages
	// received. Enabling message authentication adds significant
	// overhead for each message sent and received, so use this
	// when security outweighs performance.
	HMACKey []byte

	// TLS is a crypto/tls configuration struct. If it is present,
	// then the server will be TLS-enabled.
	TLS *tls.Config
}

// Handler is the type which functions passed to Server.Register
// must match: taking a slice of bytes as an argument and returning a
// slice of bytes and an error.
type Handler func([]byte) ([]byte, error)

// This is our dispatch table
type dispatch map[string]Handler

// New returns a new Server, ready to have handlers added.
func New(c *Config) (*Server, error) {
	if c.TLS != nil {
		l, err := tls.Listen("tcp", c.Sockname, c.TLS)
	} else {
		tcpaddr, _ := net.ResolveTCPAddr("tcp", c.Sockname)
		l, err := net.ListenTCP("tcp", tcpaddr)
	}
	if err != nil {
		return nil, err
	}
	return commonNew(c, l), nil
}

// commonNew does shared setup work for the constructors (mostly so
// that changes to Server don't have to be mirrored)
func commonNew(c *Config, l net.Listener) *Server {
	// spawn a WaitGroup and add one to it for s.sockAccept()
	var w sync.WaitGroup
	w.Add(1)
	// set c.Buffer to the default if it's zero
	if c.Buffer < 1 {
		c.Buffer = 32
	}
	// create the Server, start listening, and return
	s := &Server{make(chan *Msg, c.Buffer),
		p.Sigchan,
		make(chan bool, 1),
		c.Sockname,
		l, make(dispatch),
		time.Duration(c.Timeout) * time.Millisecond,
		c.Xferlim,
		p.Loglvl[c.Msglvl],
		c.LogIP,
		c.HMACKey,
		&w,
	}
	go s.sockAccept()
	return s
}

// Register adds a Handler function to a Server.
//
// 'name' is the command you wish this function to be the responder
// for.
//
// 'r' is the name of the Handler function which will be called on dispatch.
func (s *Server) Register(name string, r Handler) error {
	if _, ok := s.d[name]; ok {
		return fmt.Errorf("handler '%v' already exists", name)
	}
	s.d[name] = r
	return nil
}

// Quit handles shutdown and cleanup, including waiting for any
// connections to terminate. When it returns, all connections are
// fully shut down and no more work will be done.
func (s *Server) Quit() {
	s.q <- true  // send true to quit chan
	s.l.Close()  // close listener
	s.w.Wait()   // wait for waitgroup to turn down
	close(s.q)
	close(s.Msgr)
}
