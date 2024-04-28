package client // import "github.com/firepear/petrel/client"

// Copyright (c) 2014-2024 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// This file implements the Petrel client.

import (
	"crypto/tls"
	"fmt"
	"net"
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

// Client is a Petrel client instance.
type Client struct {
	conn *p.Conn
	// conn closed semaphore
	cc bool
	// transmission sequence id
	Seq uint32
}

// Config holds values to be passed to the client constructor.
type Config struct {
	// For Unix clients, Addr takes the form "/path/to/socket". For
	// TCP clients, it is either an IPv4 or IPv6 address followed by
	// the desired port number ("127.0.0.1:9090", "[::1]:9090").
	Addr string

	// Timeout is the number of milliseconds the client will wait
	// before timing out due to on a Dispatch() or Read()
	// call. Default (zero) is no timeout.
	Timeout int64

	// Xferlim is the maximum number of bytes in a single read from
	// the network. If a request exceeds this limit, the
	// connection will be dropped. Use this to prevent memory
	// exhaustion by arbitrarily long network reads. The default
	// (0) is unlimited.
	Xferlim uint32

	//HMACKey is the secret key used to generate MACs for signing
	//and verifying messages. Default (nil) means MACs will not be
	//generated for messages sent, or expected for messages
	//received.
	HMACKey []byte
}

// TCPClient returns a Client which uses TCP.
func TCPClient(c *Config) (*Client, error) {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

// TLSClient returns a Client which uses TLS + TCP.
func TLSClient(c *Config, t *tls.Config) (*Client, error) {
	conn, err := tls.Dial("tcp", c.Addr, t)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

// UnixClient returns a Client which uses Unix domain sockets.
func UnixClient(c *Config) (*Client, error) {
	conn, err := net.Dial("unix", c.Addr)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

func newCommon(c *Config, conn net.Conn) (*Client, error) {
	pconn := new(p.Conn)
	pconn.NC = conn
	pconn.Timeout = time.Duration(c.Timeout) * time.Millisecond
	pconn.Plim = c.Xferlim
	pconn.Hkey = c.HMACKey
	return &Client{pconn, false, 0}, nil
}

// Dispatch sends a request and returns the response.
func (c *Client) Dispatch(req, payload []byte) ([]byte, error) {
	c.Seq++
	c.conn.Seq = c.Seq
	// if a previous error closed the conn, refuse to do anything
	if c.cc == true {
		return nil, fmt.Errorf("the network connection is closed due to a previous error; please create a new Client")
	}
	err := p.ConnWrite(c.conn, req, payload)
	if err != nil {
		return nil, err
	}
	_, resp, err := p.ConnRead(c.conn)
	return resp, err
}

// Quit terminates the client's network connection and other
// operations.
func (c *Client) Quit() {
	c.cc = true
	c.conn.NC.Close()
}
