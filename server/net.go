package server

// Copyright (c) 2014-2024 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// Socket code for petrel

import (
	"net"

	p "github.com/firepear/petrel"
)

// sockAccept is spawned by server.commonNew. It monitors the server's
// listener socket and spawns connections for clients.
func (s *Server) sockAccept() {
	defer s.w.Done()
	var cn uint32 // connection number
	for cn = 1; true; cn++ {
		// we wait here until the listener accepts a
		// connection and spawns us a net.Conn -- or an error
		// occurs, like the listener socket closing
		nc, err := s.l.Accept()
		if err != nil {
			select {
			case <-s.q:
				// if there's a message on this
				// channel, s.Quit() was invoked and
				// we should close up shop
				s.genMsg(0, 0, 199, nil)
				return
			default:
				// otherwise, we've had an actual
				// networking error
				s.genMsg(0, 0, 599, err)
				return
			}
		}

		// if we made it down here, then we have a new
		// connection. first, wrap our net.Conn in a
		// petrel.Conn for parity with the common netcode
		pc := &p.Conn{
			NC: nc,
			Plim: s.Xferlim
			Hkey: s.HMACKey
			Timeout: time.Duration(s.Timeout) * time.Millisecond,
		}
		// increment our waitgroup
		s.w.Add(1)
		// and launch the goroutine which will actually
		// service the client
		go s.connServer(*pc, cn)
	}
}

// connServer dispatches commands from, and sends reponses to, a
// client. It is launched, per-connection, from sockAccept().
func (s *Server) connServer(c *petrel.Conn, cn uint32) {
	// queue up decrementing the waitlist and closing the network
	// connection
	defer s.w.Done()
	defer c.Quit()
	// request id for this connection
	var reqid uint32
	var response []byte

	if s.li {
		s.genMsg(cn, reqid, p.Stats["connect"], c.RemoteAddr().String(), nil)
	} else {
		s.genMsg(cn, reqid, p.Stats["connect"], "", nil)
	}

	for {
		// let us forever enshrine the dumbness of the
		// original design of the network read/write
		// functions, that we may never see their like again:
		//
		// req, payload, perr, xtra, err := p.ConnRead(c, s.t, s.rl, s.hk, &reqid)
		// perr, err = p.ConnWrite(c, req, p.Stats[perr].Xmit, s.hk, s.t, reqid)


		// read the request
		req, payload, perr, xtra, err := p.ConnRead(c, s.t, s.rl, s.hk, &reqid)
		if perr != "" {
			s.genMsg(cn, reqid, p.Stats[perr], xtra, err)
			if p.Stats[perr].Xmit != nil {
				perr, err = p.ConnWrite(c, req, p.Stats[perr].Xmit, s.hk, s.t, reqid)
				if err != nil {
					s.genMsg(cn, reqid, p.Stats[perr], "", err)
					return
				}
			}
			return
		}

		// send error if we don't recognize the command
		handler, ok := s.d[string(req)]
		if !ok {
			perr = "badreq"
			goto HANDLEERR
		}
		// dispatch the request and get the response
		s.genMsg(cn, reqid, p.Stats["dispatch"], string(req), nil)
		response, err = handler(c.Resp.Payload)
		if err != nil {
			perr = "reqerr"
			goto HANDLEERR
		}
		// send response
		perr, err = p.ConnWrite(c, req, response, s.hk, s.t, reqid)
		if err != nil {
			goto HANDLEERR
		}
		s.genMsg(cn, reqid, p.Stats["success"], "", nil)
		continue

	HANDLEERR:
		s.genMsg(cn, reqid, p.Stats[perr], string(req), err)
		if p.Stats[perr].Xmit != nil {
			perr, err = p.ConnWrite(c, req, p.Stats[perr].Xmit, s.hk, s.t, reqid)
			if err != nil {
				s.genMsg(cn, reqid, p.Stats[perr], "", err)
			}
		}
	}
}
