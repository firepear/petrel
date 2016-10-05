package petrel

// Copyright (c) 2015-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// This file implements the Petrel client.

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

// Conn is an connection to a remote service.
type Conn struct {
	conn net.Conn
	// message length buffer
	b0 []byte
	// where we read to
	b1 []byte
	// where reads accumulate
	b2 []byte
	// timeout length
	to time.Duration
	// unpacked message length
	mlen int32
	// do or do not send message len header
	prefix bool
	// bytes read from network
	bread int32
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
	Timeout int64

	// OmitPrefix, when true, causes the standard length header to be
	// omitted from dispatched messages.
	OmitPrefix bool
}

// NewTCP returns a Conn which uses TCP.
func NewTCP(c *Config) (*Conn, error) {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

// NewTLS returns a Conn which uses TLS + TCP.
func NewTLS(c *Config, t *tls.Config) (*Conn, error) {
	conn, err := tls.Dial("tcp", c.Addr, t)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

// NewUnix returns a Conn which uses Unix domain sockets.
func NewUnix(c *Config) (*Conn, error) {
	conn, err := net.Dial("unix", c.Addr)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

func newCommon(c *Config, conn net.Conn) (*Conn, error) {
	return &Conn{conn, make([]byte, 4), make([]byte, 128), nil,
		time.Duration(c.Timeout) * time.Millisecond,
		0, !c.OmitPrefix, 0}, nil
}

// Dispatch sends a request and returns the response. If Dispatch
// fails on write, call again. If it fails on read, call
// client.Read().
func (c *Conn) Dispatch(req []byte) ([]byte, error) {
	// generate packed message length header & prepend to request
	if c.prefix {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, int32(len(req)))
		req = append(buf.Bytes(), req...)
	}
	// send request
	if c.to > 0 {
		c.conn.SetDeadline(time.Now().Add(c.to))
	}
	_, err := c.conn.Write(req)
	if err != nil {
		return nil, err
	}
	if c.to > 0 {
		c.conn.SetDeadline(time.Now().Add(c.to))
	}
	resp, err := c.Read()
	return resp, err
}

// Read reads from the client connection.
func (c *Conn) Read() ([]byte, error) {
	// zero our byte-collectors
	c.b1 = make([]byte, 128)
	c.b2 = c.b2[:0]
	c.bread = 0

	// get the response message length
	if c.to > 0 {
		c.conn.SetDeadline(time.Now().Add(c.to))
	}
	n, err := c.conn.Read(c.b0)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if  n != 4 {
		return nil, fmt.Errorf("too few bytes (%v) in message length on read\n", n)
	}
	buf := bytes.NewReader(c.b0)
	err = binary.Read(buf, binary.BigEndian, &c.mlen)
	if err != nil {
		return nil, fmt.Errorf("could not decode message length on read: %v\n", err)
	}

	for c.bread < c.mlen {
		// if there are less than 128 bytes remaining to read in this
		// message, resize b1 to fit. this avoids reading across a
		// message boundary.
		if x := c.mlen - c.bread; x < 128 {
			c.b1 = make([]byte, x)
		}
		if c.to > 0 {
			c.conn.SetDeadline(time.Now().Add(c.to))
		}
		n, err = c.conn.Read(c.b1)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		c.bread += int32(n)
		c.b2 = append(c.b2, c.b1[:n]...)
	}
	// check for/handle error responses
	if c.mlen == 11 && c.b2[0] == 80 { // 11 bytes, starting with 'P'
		pp := string(c.b2[0:8])
		if pp == "PERRPERR" {
			code, err := strconv.Atoi(string(c.b2[8:11]))
			if err != nil {
				return nil, fmt.Errorf("request error: unknown code %d", code)
			}
			return nil, perrs[perrmap[code]]
		}
	}

	return c.b2[:c.mlen], err
}

// Close closes the client's connection.
func (c *Conn) Close() {
	c.conn.Close()
}
