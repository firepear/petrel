package client // import "firepear.net/asock/client"

// Copyright (c) 2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"net"
)

type Aclient struct {
	conn net.Conn
	b1   []byte
	b2   []byte
}

// NewTCP returns an asock client with a TCP connection to an asock
// instance. It takes one argument, an "address:port" string.
func NewTCP(addr string) (*Aclient, error) {
	conn, err := net.Dial("unix", sn)
	if err != nil {
		return nil, err
	}
	return &Aclient{conn, make([]byte, 64), nil}, nil
}

// Dispatch causes the client to send a request and wait for response.
func (c *Aclient) Dispatch(payload []byte) ([]byte, error) {
	c.b2 = c.b2[:0] // reslice b2 to zero it
	for {
		n, err := c.conn.Read(c.b1)
		if err != nil {
			return nil, err
		}
		c.b2 = append(c.b2, c.b1[:n]...)
		if n == 64 {
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
