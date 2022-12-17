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
				s.genMsg(0, 0, p.Errs["quit"], "", nil)
				return
			default:
				// we've had a networking error
				s.genMsg(0, 0, p.Errs["listenerfail"], "", err)
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

	if s.li {
		s.genMsg(cn, reqid, p.Errs["connect"], c.RemoteAddr().String(), nil)
	} else {
		s.genMsg(cn, reqid, p.Errs["connect"], "", nil)
	}

	for {
		// read the request
		req, rargs, perr, xtra, err := p.ConnRead(c, s.t, s.rl, s.hk, &reqid)
		if perr != "" {
			s.genMsg(cn, reqid, p.Errs[perr], xtra, err)
			if p.Errs[perr].Xmit != nil {
				perr, err = p.ConnWrite(c, req, p.Errs[perr].Xmit, s.hk, s.t, reqid)
				if err != nil {
					s.genMsg(cn, reqid, p.Errs[perr], "", err)
					return
				}
			}
			return
		}
		if len(rargs) == 0 {
			s.genMsg(cn, reqid, p.Errs["nilreq"], "", nil)
			perr, err = p.ConnWrite(c, req, p.Errs["nilreq"].Xmit, s.hk, s.t, reqid)
			if err != nil {
				s.genMsg(cn, reqid, p.Errs[perr], "", err)
				return
			}
			continue
		}

		// dispatch the request and get the response
		response, perr, req, err := s.reqDispatch(c, cn, reqid, req, rargs)
		if perr != "" {
			s.genMsg(cn, reqid, p.Errs[perr], string(req), err)
			if p.Errs[perr].Xmit != nil {
				perr, err = p.ConnWrite(c, req, p.Errs[perr].Xmit, s.hk, s.t, reqid)
				if err != nil {
					s.genMsg(cn, reqid, p.Errs[perr], "", err)
					return
				}
			}
			continue
		}

		// send response
		perr, err = p.ConnWrite(c, req, response, s.hk, s.t, reqid)
		if err != nil {
			s.genMsg(cn, reqid, p.Errs[perr], "", err)
			return
		}
		s.genMsg(cn, reqid, p.Errs["success"], "", nil)
	}
}

// reqDispatch turns the request into a command and arguments, and
// dispatches these components to a handler.
func (s *Server) reqDispatch(c net.Conn, cn, reqid uint32, req, rargs []byte) ([]byte, string, []byte, error) {
	/*
	// get chunk locations
	cl := qsplit.LocationsOnce(req)
	dcmd := string(req[cl[0]:cl[1]])
	// now get the args
	var dargs []byte
	if cl[2] != -1 {
		dargs = req[cl[2]:]
	} */
	// send error if we don't recognize the command
	responder, ok := s.d[string(req)]
	if !ok {
		return nil, "badreq", req, nil
	}
	// ok, we know the command and we have its dispatch
	// func. call it and send response
	/* var rs [][]byte // req, split by word
	switch responder.mode {
	case "argv":
		rs = qsplit.ToBytes(dargs)
	case "blob":
		rs = rs[:0]
		rs = append(rs, dargs)
	} */
	s.genMsg(cn, reqid, p.Errs["dispatch"], string(req), nil)
	response, err := responder(rargs)
	if err != nil {
		return nil, "reqerr", nil, err
	}
	return response, "", nil, nil
}
