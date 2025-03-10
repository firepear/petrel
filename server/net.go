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
	for {
		// we wait here until the listener accepts a
		// connection and spawns us a petrel.Conn -- or an
		// error occurs, like the listener socket closing
		id, sid := p.GenId()
		pc := &p.Conn{Id: id, Sid: sid, Msgr: s.Msgr}
		nc, err := s.l.Accept()
		if err != nil {
			select {
			case <-s.q:
				// if there's a message on this
				// channel, s.Quit() was invoked and
				// we should close up shop
				s.Msgr <- &p.Msg{Cid: pc.Sid, Seq: pc.Seq, Req: "NONE",
					Code: 199, Txt: "err is spurious", Err: err}
				return
			default:
				// otherwise, we've had an actual
				// networking error
				s.Msgr <- &p.Msg{Cid: pc.Sid, Seq: pc.Seq, Req: pc.Resp.Req,
					Code: 599, Txt: "unknown err", Err: err}
				return
			}
		}

		// we made it here so we have a new connection. wrap
		// our net.Conn in a petrel.Conn for parity with the
		// common netcode then add other values
		pc.NC = nc
		pc.Plim = s.rl
		pc.Hkey = s.hk
		pc.Timeout = time.Duration(s.t) * time.Millisecond

		// increment our waitgroup
		s.w.Add(1)
		// add to connlist
		s.cl.Store(id, pc)
		// and launch the goroutine which will actually
		// service the client
		go s.connServer(pc)
	}
}

// connServer dispatches commands from, and sends reponses to, a
// client. It is launched, per-connection, from sockAccept().
func (s *Server) connServer(c *p.Conn) {
	// queue up decrementing the waitlist, closing the network
	// connection, and removing the connlist entry
	defer s.w.Done()
	defer c.NC.Close()
	defer s.cl.Delete(c.Id)
	c.Msgr <- &p.Msg{Cid: c.Sid, Seq: c.Seq, Req: c.Resp.Req, Code: 100,
		Txt: fmt.Sprintf("srv:%s %s %s", s.sid, p.Stats[100].Txt,
			c.NC.RemoteAddr().String()),
		Err: nil}

	var response []byte

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
			c.Msgr <- &p.Msg{Cid: c.Sid, Seq: c.Seq, Req: c.Resp.Req,
				Code: c.Resp.Status, Txt: p.Stats[c.Resp.Status].Txt,
				Err: err}
			// don't care about err here because we're
			// gonna bail, and this may not work anyway
			_ = p.ConnWrite(c, []byte(c.Resp.Req),
				[]byte(fmt.Sprintf("%s", err)))
			break
		}
		// lookup the handler for this request
		handler, ok := s.d[c.Resp.Req]
		if ok {
			// dispatch the request and get the response
			c.Resp.Status, response, err = handler(c.Resp.Payload)
			if err != nil {
				c.Resp.Status = 500
			}
		} else {
			// unknown handler
			c.Resp.Status = 400
		}

		// we always send a response
		err = p.ConnWrite(c, []byte(c.Resp.Req), response)
		if c.Resp.Status > 1024 {
			c.Msgr <- &p.Msg{Cid: c.Sid, Seq: c.Seq, Req: c.Resp.Req,
				Code: c.Resp.Status, Txt: "app defined code", Err: err}
		} else {
			c.Msgr <- &p.Msg{Cid: c.Sid, Seq: c.Seq, Req: c.Resp.Req,
				Code: c.Resp.Status, Txt: p.Stats[c.Resp.Status].Txt,
				Err: err}
		}
		if err != nil {
			break
		}
	}
}
