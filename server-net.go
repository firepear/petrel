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
func (h *Server) sockAccept() {
	defer h.w.Done()
	var cn uint
	for cn = 1; true; cn++ {
		c, err := h.l.Accept()
		if err != nil {
			select {
			case <-h.q:
				// h.Quit() was invoked; close up shop
				h.genMsg(0, 0, perrs["quit"], "", nil)
				return
			default:
				// we've had a networking error
				h.genMsg(0, 0, perrs["listenerfail"], "", err)
				return
			}
		}
		// we have a new client
		h.w.Add(1)
		go h.connServer(c, cn)
	}
}

// connServer dispatches commands from, and sends reponses to, a client. It
// is launched, per-connection, from sockAccept().
func (h *Server) connServer(c net.Conn, cn uint) {
	defer h.w.Done()
	defer c.Close()
	// request counter for this connection
	var reqnum uint

	if h.li {
		h.genMsg(cn, reqnum, perrs["connect"], c.RemoteAddr().String(), nil)
	} else {
		h.genMsg(cn, reqnum, perrs["connect"], "", nil)
	}

	for {
		reqnum++
		// read the request
		req, perr, xtra, err := connRead(c, h.t, h.rl, h.hk)
		if perr != "" {
			h.genMsg(cn, reqnum, perrs[perr], xtra, err)
			if perrs[perr].xmit != nil {
				perr, err = connWrite(c, perrs[perr].xmit, h.hk, h.t)
				if err != nil {
					h.genMsg(cn, reqnum, perrs[perr], "", err)
					return
				}
			}
			return
		}
		if len(req) == 0 {
			h.genMsg(cn, reqnum, perrs["nilreq"], "", nil)
			perr, err = connWrite(c, perrs["nilreq"].xmit, h.hk, h.t)
			if err != nil {
				h.genMsg(cn, reqnum, perrs[perr], "", err)
				return
			}
			continue
		}

		// dispatch the request and get the reply
		reply, perr, xtra, err := h.reqDispatch(c, cn, reqnum, req)
		if perr != "" {
			h.genMsg(cn, reqnum, perrs[perr], xtra, err)
			if perrs[perr].xmit != nil {
				perr, err = connWrite(c, perrs[perr].xmit, h.hk, h.t)
				if err != nil {
					h.genMsg(cn, reqnum, perrs[perr], "", err)
					return
				}
			}
			continue
		}

		// send reply
		perr, err = connWrite(c, reply, h.hk, h.t)
		if err != nil {
			h.genMsg(cn, reqnum, perrs[perr], "", err)
			return
		}
		h.genMsg(cn, reqnum, perrs["success"], "", nil)
	}
}

// reqDispatch turns the request into a command and arguments, and
// dispatches these components to a handler.
func (h *Server) reqDispatch(c net.Conn, cn, reqnum uint, req []byte) ([]byte, string, string, error) {
	// get chunk locations
	cl := qsplit.LocationsOnce(req)
	dcmd := string(req[cl[0]:cl[1]])
	// now get the args
	var dargs []byte
	if cl[2] != -1 {
		dargs = req[cl[2]:]
	}
	// send error if we don't recognize the command
	responder, ok := h.d[dcmd]
	if !ok {
		return nil, "badreq", dcmd, nil
	}
	// ok, we know the command and we have its dispatch
	// func. call it and send response
	h.genMsg(cn, reqnum, perrs["dispatch"], dcmd, nil)
	var rs [][]byte // req, split by word
	switch responder.mode {
	case "args":
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
