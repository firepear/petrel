package petrel

// Copyright (c) 2014-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// Socket code for petrel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"firepear.net/qsplit"
)

// sockAccept monitors the listener socket and spawns connections for
// clients.
func (h *Handler) sockAccept() {
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
		go h.connHandler(c, cn)
	}
}

// connHandler dispatches commands from, and sends reponses to, a client. It
// is launched, per-connection, from sockAccept().
func (h *Handler) connHandler(c net.Conn, cn uint) {
	defer h.w.Done()
	defer c.Close()
	// request counter for this connection
	var reqnum uint

	h.genMsg(cn, reqnum, 100, c.RemoteAddr(), nil)
	for {
		reqnum++

		// read the request
		req, perr, err := h.connRead(c, cn, reqnum)
		if err != nil {
			h.genMsg(cn, reqnum, perrs[perr], "", err)
			h.send(c, cn, reqnum, perrs[perr].xmit)
			//TODO send "you've been disconnected" msg
			return
		}
		if len(req) == 0 {
			h.genMsg(cn, reqnum, perrs["nilreq"], "", nil)
			h.send(c, cn, reqnum, perrs["nilreq"].xmit)
			continue
		}

		// dispatch the request and get the reply
		reply, err := h.reqDispatch(c, cn, reqnum, req)
		if err != nil {
			continue
		}

		// send reply
		err = h.send(c, cn, reqnum, reply)
		if err != nil {
			return
		}
		h.genMsg(cn, reqnum, 200, All, "reply sent", nil)
	}
}

// connRead does all network reads and assembles the request. If it
// returns an error, then the connection terminates because the state
// of the connection cannot be known.
func (h *Handler) connRead(c net.Conn, cn, reqnum uint) ([]byte, string, error) {
	// buffer 0 holds the message length
	b0 := make([]byte, 4)
	// buffer 1: network reads go here, 128B at a time
	b1 := make([]byte, 128)
	// buffer 2: data accumulates here; requests pulled from here
	var b2 []byte
	// message length
	var mlen int32
	// bytes read so far
	var bread int32

	// get the response message length
	if h.t > 0 {
		c.SetReadDeadline(time.Now().Add(h.t))
	}
	n, err := c.Read(b0)
	if err != nil {
		if err == io.EOF {
			h.genMsg(cn, reqnum, 198, Conn, "client disconnected", err)
		} else {
			h.genMsg(cn, reqnum, 196, Conn, "failed to read mlen from socket", err)
		}
		return nil, "", err
	}
	if  n != 4 {
		h.genMsg(cn, reqnum, 196, Conn, "short read on message length", err)
		return nil, "", errshortread
	}
	buf := bytes.NewReader(b0)
	err = binary.Read(buf, binary.BigEndian, &mlen)
	if err != nil {
		h.genMsg(cn, reqnum, 501, Conn, "could not decode message length", err)
		return nil, "", err
	}

	for bread < mlen {
		// if there are less than 128 bytes remaining to read in this
		// message, resize b1 to fit. this avoids reading across a
		// message boundary.
		if x := mlen - bread; x < 128 {
			b1 = make([]byte, x)
		}
		if h.t > 0 {
			c.SetReadDeadline(time.Now().Add(h.t))
		}
		n, err = c.Read(b1)
		if err != nil {
			if err == io.EOF {
				h.genMsg(cn, reqnum, 198, Conn, "client disconnected", err)
			} else {
				h.genMsg(cn, reqnum, 196, Conn, "failed to read req from socket", err)
			}
			return nil, "", err
		}
		if n == 0 {
			// short-circuit just in case this ever manages to happen
			return b2[:mlen], "", err
		}
		bread += int32(n)
		if h.rl > 0 && bread > h.rl {
			return nil, "reqlen", perrs["reqlen"].err
		}
		b2 = append(b2, b1[:n]...)
	}
	return b2[:mlen], "", err
}

// reqDispatch turns the request into a command and arguments, and
// dispatches these components to a handler.
func (h *Handler) reqDispatch(c net.Conn, cn, reqnum uint, req []byte) ([]byte, error) {
	// get chunk locations
	cl := qsplit.LocationsOnce(req)
	dcmd := string(req[cl[0]:cl[1]])
	// now get the args
	var dargs []byte
	if cl[2] != -1 {
		dargs = req[cl[2]:]
	}
	// send error if we don't recognize the command
	dfunc, ok := h.d[dcmd]
	if !ok {
		h.send(c, cn, reqnum, []byte(fmt.Sprintf("Unknown command '%s'.", dcmd)))
		h.genMsg(cn, reqnum, 400, All, fmt.Sprintf("bad command '%s'", dcmd), nil)
		return nil, errbadcmd
	}
	// ok, we know the command and we have its dispatch
	// func. call it and send response
	h.genMsg(cn, reqnum, 101, All, fmt.Sprintf("dispatching [%s]", dcmd), nil)
	var rs [][]byte // req, split by word
	switch dfunc.mode {
	case "args":
		rs = qsplit.ToBytes(dargs)
	case "blob":
		rs = rs[:0]
		rs = append(rs, dargs)
	}
	resp, err := dfunc.df(rs)
	if err != nil {
		h.genMsg(cn, reqnum, 500, Error, "request failed", err)
		h.send(c, cn, reqnum, []byte("Sorry, an error occurred and your request could not be completed."))
		return nil, errcmderr
	}
	return resp, nil
}

// send handles all network writes.
func (h *Handler) send(c net.Conn, cn, reqnum uint, resp []byte) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, int32(len(resp)))
	if err != nil {
		h.genMsg(cn, reqnum, 501, Conn, "could not encode message length", err)
		return err
	}
	resp = append(buf.Bytes(), resp...)
	if h.t > 0 {
		c.SetReadDeadline(time.Now().Add(h.t))
	}
	_, err = c.Write(resp)
	if err != nil {
		h.genMsg(cn, reqnum, 196, Error, "failed to write resp to socket", err)
	}
	return err
}
