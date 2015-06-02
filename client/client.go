// Package client implements a basic, synchronous asock client.
package client // import "firepear.net/asock/client"

// Copyright (c) 2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"bytes"
	"crypto/tls"
	"net"
	"time"
)

// Aclient is an Asock client.
type Aclient struct {
	conn net.Conn
	b1   []byte         // where we read to
	b2   []byte         // where reads accumulate
	to   time.Duration  // I/O timeout
	eom  []byte
}

// Config holds values to be passed to the constructor.
type Config struct {
	// For Unix clients, Addr takes the form "/path/to/socket". For
	// TCP clients, it is either an IPv4 or IPv6 address followed by
	// the desired port number ("127.0.0.1:9090", "[::1]:9090").
	Addr string

	// Timeout is the number of milliseconds the client will wait
	// before timing out due to on a Dispatch() or Read()
	// call. Default (zero) is no timeout.
	Timeout time.Duration

	// EOM is the end-of-message marker. Data is read from the server
	// until EOM is encountered. Defaults to "\n\n".
	EOM string

	// TLSConfig is the configuration for TLS connections. Required
	// for NewTLS(); can be nil for all other cases.
	TLSConfig *tls.Config
}

// NewTCP returns an asock client with a TCP connection to an asock
// instance.
func NewTCP(c Config) (*Aclient, error) {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

// NewTLS returns an asock client with a TLS-secured connection to an
// asock instance.
func NewTLS(c Config) (*Aclient, error) {
	conn, err := tls.Dial("tcp", c.Addr, c.TLSConfig)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

// NewUnix returns an asock client with a Unix domain socket
// connection to an asock instance.
func NewUnix(c Config) (*Aclient, error) {
	conn, err := net.Dial("unix", c.Addr)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

func newCommon(c Config, conn net.Conn) (*Aclient, error) {
	if c.EOM == "" {
		c.EOM = "\n\n"
	}
	return &Aclient{conn, make([]byte, 128), nil, c.Timeout, []byte(c.EOM)}, nil
}

// Dispatch sends a request and returns the response. If Dispatch
// fails on write, call again. If it fails on read, call
// client.Read().
func (c *Aclient) Dispatch(req []byte) ([]byte, error) {
	if c.to > 0 {
		c.conn.SetDeadline(time.Now().Add(c.to * time.Millisecond))
	}
	req = append(req, c.eom...)
	_, err := c.conn.Write(req)
	if err != nil {
		return nil, err
	}
	return c.Read()
}

// Read reads from the client connection.
func (c *Aclient) Read() ([]byte, error) {
	var eom int
	for {
		if c.to > 0 {
			c.conn.SetDeadline(time.Now().Add(c.to * time.Millisecond))
		}
		// read off the interface into c.b1
		n, err := c.conn.Read(c.b1)
		if err != nil {
			return nil, err
		}
		// accumulate c.b1 into c.b2
		c.b2 = append(c.b2, c.b1[:n]...)
		// scan c.b2 for eom; break from loop when we find it.
		eom = bytes.Index(c.b2, c.eom)
		if eom != -1 {
			break
		}
	}
	// capture the response and reslice b2 to remove it
	// leaving b2 intact like this paves the way for async client
	// support in the future.
	resp := c.b2[:eom]
	c.b2 = c.b2[eom + len(c.eom):]
	// return the response
	return resp, nil
}

// Close closes the client's connection.
func (c *Aclient) Close() {
	c.conn.Close()
}
