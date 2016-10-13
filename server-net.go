package petrel

// Copyright (c) 2014-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// Socket code for petrel

import (
	//"crypto/hmac"
	"net"

	"firepear.net/qsplit"
)

// sockAccept monitors the listener socket and spawns connections for
// clients.
func (s *Server) sockAccept() {
	defer s.w.Done()
	var cn uint
	for cn = 1; true; cn++ {
		c, err := s.l.Accept()
		if err != nil {
			select {
			case <-s.q:
				// s.Quit() was invoked; close up shop
				s.genMsg(0, 0, perrs["quit"], "", nil)
				return
			default:
				// we've had a networking error
				s.genMsg(0, 0, perrs["listenerfail"], "", err)
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
func (s *Server) connServer(c net.Conn, cn uint) {
	defer s.w.Done()
	defer c.Close()
	// request counter for this connection
	var reqnum uint

	if s.li {
		s.genMsg(cn, reqnum, perrs["connect"], c.RemoteAddr().String(), nil)
	} else {
		s.genMsg(cn, reqnum, perrs["connect"], "", nil)
	}

	for {
		reqnum++
		// read the request
		req, perr, xtra, err := connRead(c, s.t, s.rl, s.hk)
		if perr != "" {
			s.genMsg(cn, reqnum, perrs[perr], xtra, err)
			if perrs[perr].xmit != nil {
				perr, err = connWrite(c, perrs[perr].xmit, s.hk, s.t)
				if err != nil {
					s.genMsg(cn, reqnum, perrs[perr], "", err)
					return
				}
			}
			return
		}
		if len(req) == 0 {
			s.genMsg(cn, reqnum, perrs["nilreq"], "", nil)
			perr, err = connWrite(c, perrs["nilreq"].xmit, s.hk, s.t)
			if err != nil {
				s.genMsg(cn, reqnum, perrs[perr], "", err)
				return
			}
			continue
		}

		// dispatch the request and get the reply
		reply, perr, xtra, err := s.reqDispatch(c, cn, reqnum, req)
		if perr != "" {
			s.genMsg(cn, reqnum, perrs[perr], xtra, err)
			if perrs[perr].xmit != nil {
				perr, err = connWrite(c, perrs[perr].xmit, s.hk, s.t)
				if err != nil {
					s.genMsg(cn, reqnum, perrs[perr], "", err)
					return
				}
			}
			continue
		}

		// send reply
		perr, err = connWrite(c, reply, s.hk, s.t)
		if err != nil {
			s.genMsg(cn, reqnum, perrs[perr], "", err)
			return
		}
		s.genMsg(cn, reqnum, perrs["success"], "", nil)
	}
}

// reqDispatch turns the request into a command and arguments, and
// dispatches these components to a handler.
func (s *Server) reqDispatch(c net.Conn, cn, reqnum uint, req []byte) ([]byte, string, string, error) {
	// get chunk locations
	cl := qsplit.LocationsOnce(req)
	dcmd := string(req[cl[0]:cl[1]])
	// now get the args
	var dargs []byte
	if cl[2] != -1 {
		dargs = req[cl[2]:]
	}
	// send error if we don't recognize the command
	responder, ok := s.d[dcmd]
	if !ok {
		return nil, "badreq", dcmd, nil
	}
	// ok, we know the command and we have its dispatch
	// func. call it and send response
	s.genMsg(cn, reqnum, perrs["dispatch"], dcmd, nil)
	var rs [][]byte // req, split by word
	switch responder.mode {
	case "argv":
		rs = qsplit.ToBytes(dargs)
	case "blob":
		rs = rs[:0]
		rs = append(rs, dargs)
	}
	response, err := responder.r(rs)
	if err != nil {
		return nil, "reqerr", "", err
	}
	return response, "", "", nil
}
