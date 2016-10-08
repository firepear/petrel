package petrel

// Copyright (c) 2014-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// Socket code for petrel

import (
	"bytes"
	//"crypto/hmac"
	"encoding/binary"
	"net"
	"time"

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
		req, perr, xtra, err := connRead(c, h.t, h.rl) // h.readReq(c)
		if perr != "" {
			h.genMsg(cn, reqnum, perrs[perr], xtra, err)
			if perrs[perr].xmit != nil {
				err = h.send(c, cn, reqnum, perrs[perr].xmit)
				if err != nil {
					return
				}
			}
			//TODO send "you've been disconnected" msg
			return
		}
		if len(req) == 0 {
			h.genMsg(cn, reqnum, perrs["nilreq"], "", nil)
			err = h.send(c, cn, reqnum, perrs["nilreq"].xmit)
			if err != nil {
				return
			}
			continue
		}

		// dispatch the request and get the reply
		reply, perr, xtra, err := h.reqDispatch(c, cn, reqnum, req)
		if perr != "" {
			h.genMsg(cn, reqnum, perrs[perr], xtra, err)
			if perrs[perr].xmit != nil {
				err = h.send(c, cn, reqnum, perrs[perr].xmit)
				if err != nil {
					return
				}
			}
			continue
		}

		// send reply
		err = h.send(c, cn, reqnum, reply)
		if err != nil {
			return
		}
		h.genMsg(cn, reqnum, perrs["success"], "", nil)
	}
}

// readReq does all network reads and assembles the request. If it
// returns an error, then the connection terminates because the state
// of the connection cannot be known.
//func (h *Server) readReq(c net.Conn) ([]byte, string, string, error) {
//	return connRead(c, h.t, h.rl)
//}

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

// send handles all network writes.
func (h *Server) send(c net.Conn, cn, reqnum uint, resp []byte) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, int32(len(resp)))
	if err != nil {
		h.genMsg(cn, reqnum, perrs["internalerr"], "could not encode message length", err)
		return err
	}
	resp = append(buf.Bytes(), resp...)
	if h.t > 0 {
		c.SetReadDeadline(time.Now().Add(h.t))
	}
	_, err = c.Write(resp)
	if err != nil {
		h.genMsg(cn, reqnum, perrs["netwriteerr"], "", err)
		return err
	}
	return err
}
