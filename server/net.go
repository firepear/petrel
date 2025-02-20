package server

// Copyright (c) 2014-2025 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// Socket code for petrel

import (
	"fmt"
	"time"

	p "github.com/firepear/petrel"
)

// sockAccept is spawned by server.commonNew. It monitors the server's
// listener socket and spawns connections for clients.
func (s *Server) sockAccept() {
	defer s.w.Done()
	var cn uint32 // connection number
	for cn = 1; true; cn++ {
		// we wait here until the listener accepts a
		// connection and spawns us a petrel.Conn -- or an
		// error occurs, like the listener socket closing
		pc := &p.Conn{Id: cn, Msgr: s.Msgr}
		nc, err := s.l.Accept()
		if err != nil {
			select {
			case m := <-s.q:
				// if there's a message on this
				// channel, s.Quit() was invoked and
				// we should close up shop
				pc.GenMsg(199, pc.Resp.Req, fmt.Errorf("%v", m))
				pc.GenMsg(199, pc.Resp.Req, err)
				return
			default:
				// otherwise, we've had an actual
				// networking error
				pc.GenMsg(599, pc.Resp.Req, err)
				return
			}
		}

		// we made it here so we have a new connection. wrap
		// our net.Conn in a petrel.Conn for parity with the
		// common netcode then add other values
		pc.NC = nc
		pc.ML = s.ml
		pc.Plim = s.rl
		pc.Hkey = s.hk
		pc.Timeout = time.Duration(s.t) * time.Millisecond

		// increment our waitgroup
		s.w.Add(1)
		// and launch the goroutine which will actually
		// service the client
		go s.connServer(pc, cn)
	}
}

// connServer dispatches commands from, and sends reponses to, a
// client. It is launched, per-connection, from sockAccept().
func (s *Server) connServer(c *p.Conn, cn uint32) {
	// queue up decrementing the waitlist and closing the network
	// connection
	defer s.w.Done()
	defer c.NC.Close()

	var response []byte

	if s.li {
		c.GenMsg(100, c.Resp.Req, fmt.Errorf("%s", c.NC.RemoteAddr().String()))
	} else {
		c.GenMsg(100, c.Resp.Req, nil)
	}

	for {
		// let us forever enshrine the dumbness of the
		// original design of the network read/write
		// functions, that we may never see their like again:
		//
		// req, payload, perr, xtra, err := p.ConnRead(c, s.t, s.rl, s.hk, &reqid)
		// perr, err = p.ConnWrite(c, req, p.Stats[perr].Xmit, s.hk, s.t, reqid)

		// read the request
		err := p.ConnRead(c)
		if err != nil || c.Resp.Status > 399 {
			c.GenMsg(c.Resp.Status, c.Resp.Req, err)
			return
		}
		// lookup the handler for this request
		handler, ok := s.d[c.Resp.Req]
		if ok {
			// dispatch the request and get the response
			c.Resp.Status, response, err = handler(c.Resp.Payload)
			if err != nil {
				c.Resp.Status = 500
				c.GenMsg(c.Resp.Status, c.Resp.Req, err)
				continue
			}
		} else {
			// unknown handler
			c.Resp.Status = 400
		}
		// send response
		err = p.ConnWrite(c, []byte(c.Resp.Req), response)
		if err != nil {
			c.GenMsg(c.Resp.Status, c.Resp.Req, err)
			return
		}
		c.GenMsg(c.Resp.Status, c.Resp.Req, err)
	}
}
