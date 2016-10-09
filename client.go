package petrel

// Copyright (c) 2015-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// This file implements the Petrel client.

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"time"
)

// Client is a Petrel client instance.
type Client struct {
	conn net.Conn
	// timeout length
	to time.Duration
	// HMAC key
	hk []byte
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
	return &Client{conn, time.Duration(c.Timeout) * time.Millisecond, c.HMACKey}, nil
}

// Dispatch sends a request and returns the response.
func (c *Client) Dispatch(req []byte) ([]byte, error) {
	_, err := connWrite(c.conn, req, c.hk, c.to)
	if err != nil {
		return nil, err
	}
	resp, err := c.read()
	return resp, err
}

// read reads from the network.
func (c *Client) read() ([]byte, error) {
	resp, perr, _, err := connRead(c.conn, c.to, 0, c.hk)
	if err != nil {
		return nil, err
	}
	if perr != "" {
		return nil, perrs[perr]
	}
	// check for/handle remote-side error responses
	if len(resp) == 11 && resp[0] == 80 { // 11 bytes, starting with 'P'
		pp := string(resp[0:8])
		if pp == "PERRPERR" {
			code, err := strconv.Atoi(string(resp[8:11]))
			if err != nil {
				return []byte{255}, fmt.Errorf("request error: unknown code %d", code)
			}
			return []byte{255}, perrs[perrmap[code]]
		}
	}
	return resp, err
}

// Close closes the client's connection.
func (c *Client) Close() {
	c.conn.Close()
}
