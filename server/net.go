package server

// Copyright (c) 2014-2022 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// Socket code for petrel

import (
	"net"

	p "github.com/firepear/petrel"
)

// sockAccept monitors the listener socket and spawns connections for
// clients.
func (s *Server) sockAccept() {
	defer s.w.Done()
	var cn uint32
	for cn = 1; true; cn++ {
		c, err := s.l.Accept()
		if err != nil {
			select {
			case <-s.q:
				// s.Quit() was invoked; close up shop
				s.genMsg(0, 0, p.Stats["quit"], "", nil)
				return
			default:
				// we've had a networking error
				s.genMsg(0, 0, p.Stats["listenerfail"], "", err)
				return
			}
		}
		// we have a new client
		s.w.Add(1)
		go s.connServer(c, cn)
	}
}

// connServer dispatches commands from, and sends reponses to, a client. It
// is launched, per-connection, from sockAccept().
func (s *Server) connServer(c net.Conn, cn uint32) {
	defer s.w.Done()
	defer c.Close()
	// request id for this connection
	var reqid uint32
	var response []byte

	if s.li {
		s.genMsg(cn, reqid, p.Stats["connect"], c.RemoteAddr().String(), nil)
	} else {
		s.genMsg(cn, reqid, p.Stats["connect"], "", nil)
	}

	for {
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
		responder, ok := s.d[string(req)]
		if !ok {
			perr = "badreq"
			goto HANDLEERR
		}
		// dispatch the request and get the response
		s.genMsg(cn, reqid, p.Stats["dispatch"], string(req), nil)
		response, err = responder(payload)
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
