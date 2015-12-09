package asock

// Copyright (c) 2014,2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// Socket code for asock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"firepear.net/qsplit"
)

// sockAccept monitors the listener socket and spawns connections for
// clients.
func (a *Asock) sockAccept() {
	defer a.w.Done()
	var cn uint
	for cn = 1; true; cn++ {
		c, err := a.l.Accept()
		if err != nil {
			select {
			case <-a.q:
				// a.Quit() was invoked; close up shop
				a.Msgr <- &Msg{0, 0, 199, "Quit called: closing listener socket", nil}
				return
			default:
				// we've had a networking error
				a.Msgr <- &Msg{0, 0, 599, "read from listener socket failed", err}
				return
			}
		}
		// we have a new client
		a.w.Add(1)
		go a.connHandler(c, cn)
	}
}

// connHandler dispatches commands from, and sends reponses to, a client. It
// is launched, per-connection, from sockAccept().
func (a *Asock) connHandler(c net.Conn, cn uint) {
	defer a.w.Done()
	defer c.Close()
	// a request, split by word
	var rs [][]byte
	// request counter for this connection
	var reqnum uint

	a.genMsg(cn, reqnum, 100, Conn, "client connected", nil)
	for {
		reqnum++
		// check if we're a one-shot connection, and if we're done
		if a.t < 0 && reqnum > 1 {
			a.genMsg(cn, reqnum, 197, Conn, "ending session", nil)
			return
		}
		// read the request
		req, err := a.connRead(c, cn)
		// extract the command
		if len(req) == 0 {
			c.Write([]byte(fmt.Sprintf("Received empty request. Available commands: %v", a.help)))
			a.genMsg(cn, reqnum, 401, All, "nil request", nil)
			continue
		}
		cl := qsplit.Locations(req)
		dcmd := string(req[cl[0][0]:cl[0][1]])
		// now get the args
		var dargs []byte
		if len(cl) == 1 {
			dargs = nil
		} else {
			dargs = req[cl[1][0]:]
		}
		// send error and list of known commands if we don't
		// recognize the command
		dfunc, ok := a.d[dcmd]
		if !ok {
			c.Write([]byte(fmt.Sprintf("Unknown command '%s'. Available commands: %s",
				dcmd, a.help)))
			a.genMsg(cn, reqnum, 400, All, fmt.Sprintf("bad command '%s'", dcmd), nil)
			continue
		}
		// ok, we know the command and we have its dispatch
		// func. call it and send response
		a.genMsg(cn, reqnum, 101, All, fmt.Sprintf("dispatching [%s]", dcmd), nil)
		switch dfunc.argmode {
		case "split":
			rs = qsplit.ToBytes(dargs)
		case "nosplit":
			rs = rs[:0]
			rs = append(rs, dargs)
		}
		reply, err := dfunc.df(rs)
		if err != nil {
			c.Write([]byte("Sorry, an error occurred and your request could not be completed."))
			a.genMsg(cn, reqnum, 500, Error, "request failed", err)
			continue
		}
		c.Write(reply)
		a.genMsg(cn, reqnum, 200, All, "reply sent", nil)
	}
}

func (a *Asock) connRead(c net.Conn, cn uint) ([]byte, error) {
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

	// zero our byte-collectors; set timeout deadline
	b1 = b1[:0]
	b2 = b2[:0]
	bread = 0
	a.setConnTimeout(c)

	// get the response message length
	n, err := c.Read(b0)
	if err != nil {
		return nil, err
	}
	if  n != 4 {
		return nil, fmt.Errorf("too few bytes (%v) in message length on read: %v\n", n)
	}
	buf := bytes.NewReader(b0)
	err = binary.Read(buf, binary.BigEndian, &mlen)
	if err != nil {
		return nil, fmt.Errorf("could not decode message length on read: %v\n", err)
	}

	for bread < mlen {
		if x := mlen - bread; x < 128 {
			b1 = b1[:0]
		}
		a.setConnTimeout(c)
		n, err := c.Read(b1)
		if err != nil && err.Error() != "EOF" {
			return nil, err
		}
		bread += int32(n)
		b2 = append(b2, b1[:n]...)
	}
	return b2[:mlen], nil
}

/* old conn read code
		// set conn timeout deadline if needed
		if a.t != 0 {
		}
		// get some data from the client
		b, err := c.Read(b1)
		if err != nil {
			if err.Error() == "EOF" {
				a.genMsg(cn, reqnum, 198, Conn, "client disconnected", err)
			} else {
				a.genMsg(cn, reqnum, 197, Conn, "ending session", err)
			}
			return
		}
		// append what we read into the b2 slice
		b2 = append(b2, b1[:b]...)
		// and enter the dispatch loop
		for {
			// scan b2 for eom; break from loop if we don't find it.
			eom := bytes.Index(b2, a.eom)
			if eom == -1 {
				break
			}
*/

func (a *Asock) setConnTimeout(c net.Conn) {
	var t time.Duration
	if a.t > 0 {
		t = time.Duration(a.t)
	} else {
		t = time.Duration(a.t - (a.t * 2))
	}
	_ = c.SetReadDeadline(time.Now().Add(t * time.Millisecond))
}
