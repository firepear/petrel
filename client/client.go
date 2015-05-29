// Package client implements a basic, synchronous asock client.
package client // import "firepear.net/asock/client"

// Copyright (c) 2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	//"bytes"
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

type Config struct {
	Addr string
	Timeout time.Duration
	EOM string
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
func (c *Aclient) Dispatch(request []byte) ([]byte, error) {
	if c.to > 0 {
		c.conn.SetDeadline(time.Now().Add(c.to * time.Millisecond))
	}
	request = append(request, c.eom...)
	_, err := c.conn.Write(request)
	if err != nil {
		return nil, err
	}
	return c.Read()
}

// Read reads from the client connection.
func (c *Aclient) Read() ([]byte, error) {
	c.b2 = c.b2[:0] // reslice b2 to zero it
	for {
		if c.to > 0 {
			c.conn.SetDeadline(time.Now().Add(c.to * time.Millisecond))
		}
		n, err := c.conn.Read(c.b1)
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
