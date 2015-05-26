// Package client implements a basic, synchronous asock client.
package client // import "firepear.net/asock/client"

// Copyright (c) 2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"crypto/tls"
	"net"
	"time"
)

type Aclient struct {
	conn net.Conn
	b1   []byte
	b2   []byte
}

// NewTCP returns an asock client with a TCP connection to an asock
// instance. It takes two arguments: an "address:port" string, and the
// number of seconds until socket read/write ops timeout (0 for no
// timeout).
func NewTCP(addr string, timeout time.Duration) (*Aclient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(timeout * time.Second))
	return &Aclient{conn, make([]byte, 128), nil}, nil
}

// NewTLS returns an asock client with a TLS-secured connection to an
// asock instance. In addition to the TLS configuration, it takes an
// "address:port" argument, the number of seconds until socket
// read/write ops timeout (0 for no timeout).
func NewTLS(addr string, timeout time.Duration, tc *tls.Config) (*Aclient, error) {
	conn, err := tls.Dial("tcp", addr, tc)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(timeout * time.Second))
	return &Aclient{conn, make([]byte, 128), nil}, nil
}

// NewUnix returns an asock client with a Unix domain socket
// connection to an asock instance. It takes two arguments, a
// "/path/to/socket" string, and the number of seconds until socket
// read/write ops timeout (0 for no timeout).
func NewUnix(path string, timeout time.Duration) (*Aclient, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(timeout * time.Second))
	return &Aclient{conn, make([]byte, 128), nil}, nil
}

// Dispatch sends a request and waits for the response.
func (c *Aclient) Dispatch(request []byte) ([]byte, error) {
	n, err := c.conn.Write(request)
	if err != nil {
		return nil, err
	}
	c.b2 = c.b2[:0] // reslice b2 to zero it
	for {
		n, err = c.conn.Read(c.b1)
		if err != nil {
			return nil, err
		}
		c.b2 = append(c.b2, c.b1[:n]...)
		if n == 128 {
			continue
		}
		break
	}
	return c.b2, nil
}

// Close closes the client's connection.
func (c *Aclient) Close() {
	c.conn.Close()
}
