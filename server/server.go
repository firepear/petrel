package server

// Copyright (c) 2014-2025 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	p "github.com/firepear/petrel"
)

var (
	loglvl = map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
	}
)

// Server is a Petrel server instance.
type Server struct {
	// Msgr is the internal-facing channel which receives
	// notifications from connections.
	Msgr chan *p.Msg
	// Shutdown is the external-facing channel which notifies
	// applications that a Server instance is shutting down
	Shutdown chan error
	sig      chan os.Signal     // p.Sigchan; OS signals
	id       string             // server id
	sid      string             // short id
	q        chan bool          // quit signal socket
	s        string             // socket name
	l        net.Listener       // listener socket
	d        map[string]Handler // dispatch table
	cl       sync.Map           // connection list
	t        time.Duration      // timeout
	rl       uint32             // request length
	ml       int                // message level
	li       bool               // log ip flag
	hk       []byte             // HMAC key
	w        *sync.WaitGroup
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

	// Xferlim is the maximum number of bytes in a single read
	// from the network. If a request exceeds this limit, the
	// connection will be dropped. Use this to prevent memory
	// exhaustion by arbitrarily long network reads. The default
	// (0) is unlimited. The message header counts toward the
	// limit, so very small limits or payloads that bump up
	// against the limit may cause unexpected failures.
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

// Handler is the type which functions passed to Server.Register must
// match: taking a slice of bytes as an argument; and returning a
// uint16 (indicating status), a slice of bytes (the response), and an
// error.
//
// Petrel reserves the status range 1-2048 for internal
// use. Applications may use codes in this range, but the system will
// interpret them according to their defined meanings (e.g. it is
// standard to return '200' for success with no additional
// context). Applications are free define the remaining codes, up to
// 65535, as they see fit.
type Handler func([]byte) (uint16, []byte, error)

// New returns a new Server, ready to have handlers added.
func New(c *Config) (*Server, error) {
	var l net.Listener
	var err error

	if c.TLS != nil {
		l, err = tls.Listen("tcp", c.Sockname, c.TLS)
	} else {
		tcpaddr, _ := net.ResolveTCPAddr("tcp", c.Sockname)
		l, err = net.ListenTCP("tcp", tcpaddr)
	}
	if err != nil {
		return nil, err
	}
	return commonNew(c, l)
}

// commonNew does shared setup work for the constructors (mostly so
// that changes to Server don't have to be mirrored)
func commonNew(c *Config, l net.Listener) (*Server, error) {
	// set c.Buffer to the default if it's zero
	if c.Buffer == 0 {
		c.Buffer = 32
	}

	// generate id and short id
	id, sid := p.GenId()

	// create the Server, start listening, and return
	s := &Server{make(chan *p.Msg, c.Buffer),
		make(chan error, 4),
		p.Sigchan,
		id,
		sid,
		make(chan bool, 1),
		c.Sockname,
		l,
		make(map[string]Handler),
		sync.Map{},
		time.Duration(c.Timeout) * time.Millisecond,
		c.Xferlim,
		loglvl[c.Msglvl],
		c.LogIP,
		c.HMACKey,
		&sync.WaitGroup{},
	}

	// add one to waitgroup for s.sockAccept()
	s.w.Add(1)
	// start msgHandler event func
	go msgHandler(s)
	// launch the listener socket event func
	go s.sockAccept()
	// register the PROTOCHECK handler, called by all clients
	// during connection
	err := s.Register("PROTOCHECK", protocheck)
	// all done
	if err == nil {
		log.Printf("petrel server %s up on %s", s.sid, c.Sockname)
	}
	return s, err
}

// Register adds a Handler function to a Server.
//
// 'name' is the command you wish this function to be the responder
// for.
//
// 'r' is the name of the Handler function which will be called on dispatch.
func (s *Server) Register(name string, r Handler) error {
	if _, ok := s.d[name]; ok {
		return fmt.Errorf("handler '%s' already exists", name)
	}
	s.d[name] = r
	return nil
}

// Quit handles shutdown and cleanup, including waiting for any
// connections to terminate. When it returns, all connections are
// fully shut down and no more work will be done.
func (s *Server) Quit() {
	s.q <- true // send true to quit chan
	s.l.Close() // close listener
	s.w.Wait()  // wait for waitgroup to turn down
	close(s.q)
	close(s.Msgr)
}

// msgHandler is a function which we'll launch later on as a
// goroutine. It listens to our Server's Msgr channel, checking for a
// few critical things and logging everything else informationally.
func msgHandler(s *Server) {
	keepalive := true
	for keepalive {
		select {
		case msg := <-s.Msgr:
			switch msg.Code {
			case 599:
				// 599 is "the Server listener socket has
				// died". call s.Quit() to clean things up,
				// send the Msg to our main routine, then kill
				// this loop
				s.Shutdown <- msg
				keepalive = false
				s.Quit()
			case 199:
				// 199 is "we've been told to quit", so we
				// want to break out of the loop here as well
				s.Shutdown <- msg
				keepalive = false
			default:
				// anything else we'll log to the console to
				// show what's going on under the hood!
				log.Println(msg)
			}
		case <-s.sig:
			s.Shutdown <- fmt.Errorf("OS sig rec'd; server %s shutting down", s.sid)
			s.Quit()
		}
	}
}

// protocheck implements the mandatory protocol check handler
func protocheck(proto []byte) (uint16, []byte, error) {
	if proto[0] == p.Proto[0] {
		return 200, p.Proto, nil
	}
	return 497, p.Proto, nil
}
