package client

// Copyright (c) 2015-2022 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// This file implements the Petrel client.

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"

	p "github.com/firepear/petrel"
)

// Client is a Petrel client instance.
type Client struct {
	conn net.Conn
	// timeout length
	to time.Duration
	// HMAC key
	hk []byte
	// conn closed semaphore
	cc bool
	// transmission sequence id
	Seq uint32
}

// ClientConfig holds values to be passed to the client constructor.
type ClientConfig struct {
	// For Unix clients, Addr takes the form "/path/to/socket". For
	// TCP clients, it is either an IPv4 or IPv6 address followed by
	// the desired port number ("127.0.0.1:9090", "[::1]:9090").
	Addr string

	// Timeout is the number of milliseconds the client will wait
	// before timing out due to on a Dispatch() or Read()
	// call. Default (zero) is no timeout.
	Timeout int64

	//HMACKey is the secret key used to generate MACs for signing
	//and verifying messages. Default (nil) means MACs will not be
	//generated for messages sent, or expected for messages
	//received.
	HMACKey []byte
}

// TCPClient returns a Client which uses TCP.
func TCPClient(c *ClientConfig) (*Client, error) {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

// TLSClient returns a Client which uses TLS + TCP.
func TLSClient(c *ClientConfig, t *tls.Config) (*Client, error) {
	conn, err := tls.Dial("tcp", c.Addr, t)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

// UnixClient returns a Client which uses Unix domain sockets.
func UnixClient(c *ClientConfig) (*Client, error) {
	conn, err := net.Dial("unix", c.Addr)
	if err != nil {
		return nil, err
	}
	return newCommon(c, conn)
}

func newCommon(c *ClientConfig, conn net.Conn) (*Client, error) {
	return &Client{conn, time.Duration(c.Timeout) * time.Millisecond, c.HMACKey, false, 0}, nil
}

// Dispatch sends a request and returns the response.
func (c *Client) Dispatch(req []byte) ([]byte, error) {
	c.Seq++
	// if a previous error closed the conn, refuse to do anything
	if c.cc == true {
		return nil, fmt.Errorf("the network connection is closed due to a previous error; please create a new Client")
	}
	_, err := p.ConnWrite(c.conn, req, c.hk, c.to, c.Seq)
	if err != nil {
		return nil, err
	}
	resp, err := c.read(false)
	return resp, err
}

// DispatchRaw sends a pre-encoded transmission and returns the
// response.
func (c *Client) DispatchRaw(xmission []byte) ([]byte, error) {
	// if a previous error closed the conn, refuse to do anything
	if c.cc == true {
		return nil, fmt.Errorf("the network connection is closed due to a previous error; please create a new Client")
	}
	_, err := p.ConnWriteRaw(c.conn, c.to, xmission)
	if err != nil {
		return nil, err
	}
	resp, err := c.read(true)
	return resp, err
}

// read reads from the network.
func (c *Client) read(raw bool) ([]byte, error) {
	var resp []byte
	var perr string
	var err error
	if raw {
		resp, perr, _, err = p.ConnReadRaw(c.conn, c.to)
	} else {
		resp, perr, _, err = p.ConnRead(c.conn, c.to, 0, c.hk, &c.Seq)
	}
	if err != nil {
		return nil, err
	}
	if perr != "" {
		return nil, p.Errs[perr]
	}
	// check for/handle remote-side error responses
	if len(resp) == 11 && resp[0] == 80 { // 11 bytes, starting with 'P'
		pp := string(resp[0:8])
		if pp == "PERRPERR" {
			code, err := strconv.Atoi(string(resp[8:11]))
			if code == 402 || code == 502 {
				c.Quit()
			}
			if err != nil {
				return []byte{255}, fmt.Errorf("request error: unknown code %d", code)
			}
			return []byte{255}, p.Errs[p.Errmap[code]]
		}
	}
	return resp, err
}

// Quit terminates the client's network connection and other
// operations.
func (c *Client) Quit() {
	c.cc = true
	c.conn.Close()
}
