package asock

// Copyright (c) 2014,2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// Socket code for asock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"firepear.net/qsplit"
)

var (
	errshortread = fmt.Errorf("too few bytes")
	errbadcmd = fmt.Errorf("bad command")
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
		req, err := a.connRead(c, cn, reqnum)
		if err != nil {
			// TODO write "you're being dropped" msg
			return
		}
		if len(req) == 0 {
			a.sendMsg(c, cn, reqnum, []byte(fmt.Sprintf("Received empty request. Available commands: %v", a.help)))
			a.genMsg(cn, reqnum, 401, All, "nil request", nil)
			continue
		}

		// dispatch the request and get the reply
		reply, err := a.reqDispatch(c, cn, reqnum, req)
		if err != nil {
			continue
		}

		// send reply
		err = a.sendMsg(c, cn, reqnum, reply)
		if err != nil {
			return
		}
		a.genMsg(cn, reqnum, 200, All, "reply sent", nil)
	}
}

func (a *Asock) connRead(c net.Conn, cn, reqnum uint) ([]byte, error) {
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

	// zero our byte-accumulator; set timeout deadline
	b2 = b2[:0]
	bread = 0

	// get the response message length
	a.setConnTimeout(c)
	n, err := c.Read(b0)
	if err != nil {
		if err == io.EOF {
			a.genMsg(cn, reqnum, 198, Conn, "client disconnected", err)
		} else {
			a.genMsg(cn, reqnum, 501, Conn, "failed to read mlen from socket", err)
		}
		return nil, err
	}
	if  n != 4 {
		a.genMsg(cn, reqnum, 501, Conn, "short read on message length", err)
		return nil, errshortread
	}
	buf := bytes.NewReader(b0)
	err = binary.Read(buf, binary.BigEndian, &mlen)
	if err != nil {
		a.genMsg(cn, reqnum, 501, Conn, "could not decode message length", err)
		return nil, err
	}

	for bread < mlen {
		a.setConnTimeout(c)
		n, err := c.Read(b1)
		if err != nil {
			if err == io.EOF {
				a.genMsg(cn, reqnum, 198, Conn, "client disconnected", err)
			} else {
				a.genMsg(cn, reqnum, 501, Conn, "failed to read req from socket", err)
				return nil, err
			}
		}
		if n == 0 {
			// short-circuit just in case this ever manages to happen
			return b2[:mlen], err
		}
		bread += int32(n)
		b2 = append(b2, b1[:n]...)
	}
	return b2[:mlen], err
}

func (a *Asock) reqDispatch(c net.Conn, cn, reqnum uint, req []byte) ([]byte, error) {
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
		a.sendMsg(c, cn, reqnum, []byte(fmt.Sprintf("Unknown command '%s'. Available commands: %s", dcmd, a.help)))
		a.genMsg(cn, reqnum, 400, All, fmt.Sprintf("bad command '%s'", dcmd), nil)
		return nil, errbadcmd
	}
	// ok, we know the command and we have its dispatch
	// func. call it and send response
	a.genMsg(cn, reqnum, 101, All, fmt.Sprintf("dispatching [%s]", dcmd), nil)
	var rs [][]byte // req, split by word
	switch dfunc.argmode {
	case "split":
		rs = qsplit.ToBytes(dargs)
	case "nosplit":
		rs = rs[:0]
		rs = append(rs, dargs)
	}
	resp, err := dfunc.df(rs)
	if err != nil {
		a.genMsg(cn, reqnum, 500, Error, "request failed", err)
		err = a.sendMsg(c, cn, reqnum, []byte("Sorry, an error occurred and your request could not be completed."))
		return nil, err
	}
	return resp, nil
}

func (a *Asock) sendMsg(c net.Conn, cn, reqnum uint, resp []byte) error {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, int32(len(resp)))
	resp = append(buf.Bytes(), resp...)
	a.setConnTimeout(c)
	_, err := c.Write(resp)
	if err != nil {
		a.genMsg(cn, reqnum, 502, Error, "failed to write resp to socket", err)
	}
	return err
}

func (a *Asock) setConnTimeout(c net.Conn) {
	if a.t == 0 {
		return
	}
	var t time.Duration
	if a.t > 0 {
		t = time.Duration(a.t)
	} else {
		t = time.Duration(a.t - (a.t * 2))
	}
	_ = c.SetReadDeadline(time.Now().Add(t * time.Millisecond))
}
