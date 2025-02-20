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
	Resp *p.Resp
	conn *p.Conn
	// conn closed semaphore
	cc bool
}

// Config holds values to be passed to the client constructor.
type Config struct {
	// Address is either an IPv4 or IPv6 address followed by the
	// desired port number ("127.0.0.1:9090", "[::1]:9090").
	Addr string

	// Timeout is the number of milliseconds the client will wait
	// before timing out due to on a Dispatch() or Read()
	// call. Default is no timeout (zero).
	Timeout int64

	// TLS is the (optional) TLS configuration. If it is nil, the
	// connection will be unencrypted.
	TLS *tls.Config

	// Xferlim is the maximum number of bytes in a single read
	// from the network (functionally it limits request or
	// response payload size). If a read exceeds this limit,
	// the connection will be dropped. Use this to prevent memory
	// exhaustion by arbitrarily long network reads. The default
	// (0) is unlimited.
	Xferlim uint32

	//HMACKey is the secret key used to generate MACs for signing
	//and verifying messages. Default (nil) means MACs will not be
	//generated for messages sent, or expected for messages
	//received.
	HMACKey []byte
}

// New returns a new Client, configured and ready to use.
func New(c *Config) (*Client, error) {
	var conn net.Conn
	var err error

	if c.TLS == nil {
		conn, err = net.Dial("tcp", c.Addr)
	} else {
		conn, err = tls.Dial("tcp", c.Addr, c.TLS)
	}
	if err != nil {
		return nil, err
	}

	pconn := &p.Conn{
		NC:      conn,
		Plim:    c.Xferlim,
		Hkey:    c.HMACKey,
		Timeout: time.Duration(c.Timeout) * time.Millisecond,
	}
	client := &Client{&pconn.Resp, pconn, false}

	err = client.Dispatch("PROTOCHECK", p.Proto)
	if err != nil {
		return nil, err
	}
	if client.Resp.Status == 400 {
		return nil, fmt.Errorf("server does not support protocol version check")
	}
	if client.Resp.Status == 497 {
		return nil, fmt.Errorf("%s: client v%d; server v%d",
			p.Stats[client.Resp.Status].Txt, p.Proto[0], client.Resp.Payload[0])
	}
	if client.Resp.Status != 200 {
		client.Quit()
		return nil, fmt.Errorf("%s: %s",
			p.Stats[client.Resp.Status].Txt, client.Resp.Req)
	}
	return client, nil
}

// Dispatch sends a request and places the response in Client.Resp. If
// Resp.Status has a level of Error or Fatal, the Client will close
// its network connection
func (c *Client) Dispatch(req string, payload []byte) error {
	// if a previous error closed the conn, refuse to do anything
	if c.cc == true {
		return fmt.Errorf("network connection closed; please create a new Client")
	}
	// check for cmd length
	if len(req) > 255 {
		return fmt.Errorf("invalid request '%s': > 255 bytes", req)
	}
	// increment sequence
	c.conn.Seq++
	// send data
	err := p.ConnWrite(c.conn, []byte(req), payload)
	if err != nil {
		return fmt.Errorf("failed to send request: %s: %v",
			p.Stats[c.Resp.Status].Txt, err)
	}
	// read response
	err = p.ConnRead(c.conn)
	// if our response status is Error or Fatal, close the
	// connection and flag ourselves as done
	if p.Stats[c.Resp.Status].Lvl > p.Warn {
		c.Quit()
	}
	return err
}

func (c *Client) StatusTxt() (txt string) {
	return p.Stats[c.Resp.Status].Txt
}

// Quit terminates the client's network connection and other
// operations.
func (c *Client) Quit() {
	c.cc = true
	c.conn.NC.Close()
}
