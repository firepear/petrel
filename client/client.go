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
// instance. It takes one argument, an "address:port" string.
func NewTCP(addr string) (*Aclient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	return &Aclient{conn, make([]byte, 128), nil}, nil
}

// NewTLS returns an asock client with a TLS-secured connection to an
// asock instance. It takes an "address:port" argument.
func NewTLS(addr string, tc *tls.Config) (*Aclient, error) {
	conn, err := tls.Dial("tcp", addr, tc)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	return &Aclient{conn, make([]byte, 128), nil}, nil
}

// NewUnix returns an asock client with a Unix domain socket
// connection to an asock instance. It takes one argument, a
// "/path/to/socket" string.
func NewUnix(path string) (*Aclient, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(10 * time.Second))
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
