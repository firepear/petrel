package adminsock

// Copyright (c) 2014,2015 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

// Socket code for adminsock

import (
	"fmt"
	"net"
	"time"
	
	"firepear.net/goutils/qsplit"
)

// sockAccept monitors the listener socket and spawns connections for
// clients.
func (a *Adminsock) sockAccept() {
	defer a.w.Done()
	for n := 1; true; n++ {
		c, err := a.l.Accept()
		if err != nil {
			select {
			case <-a.q:
				// a.Quit() was invoked; close up shop
				a.genMsg(0, 0, 199, 2, "closing listener socket", nil)
				return
			default:
				// we've had a networking error
				a.genMsg(0, 0, 599, 4, "read from listener socket failed", err)
				return
			}
		}
		// we have a new client
		a.w.Add(1)
		go a.connHandler(c, n)
	}
}

// connHandler dispatches commands from, and sends reponses to, a client. It
// is launched, per-connection, from sockAccept().
func (a *Adminsock) connHandler(c net.Conn, n int) {
	defer a.w.Done()
	defer c.Close()
	b1 := make([]byte, 64) // buffer 1:  network reads go here, 64B at a time
	var b2 []byte          // buffer 2:  then are accumulated here
	var bs []string        // b2, turned into strings by word
	var reqnum int         // request counter for this connection
	var cmdhelp string     // list of commands for the auto-help msg
	for cmd := range a.d {
		cmdhelp = cmdhelp + "    " + cmd + "\n"
	}
	a.genMsg(n, reqnum, 100, 2, "client connected", nil)
	for {
		// set conn timeout deadline if needed
		if a.t != 0 {
			var t time.Duration
			if a.t > 0 {
				t = time.Duration(a.t)
			} else {
				t = time.Duration(a.t - (a.t * 2))
			}
			err := c.SetReadDeadline(time.Now().Add(t * time.Second))
			if err != nil {
				a.genMsg(n, reqnum, 501, 3, "deadline set failed; disconnecting client", err)
				c.Write([]byte("Sorry, an error occurred. Terminating connection."))
				return
			}
		}
		// get request from the client
		for {
			b, err := c.Read(b1)
			if err != nil {
				a.genMsg(n, reqnum, 197, 2, "client disconnected", err)
				return
			}
			if b > 0 {
				// then copy those bytes into the b2 slice
				b2 = append(b2, b1[:b]...)
				// if we read 64 bytes, loop back to get anything that
				// might be left on the wire
				if b == 64 {
					continue
				}
				bs = qsplit.ToStrings(b2)
				// reslice b2 so that it will be "empty" on the next read
				b2 = b2[:0]
				// break inner loop; drop to dispatch
				break 
			}
		}
		reqnum++
		// bs[0] is the command. dispatch if we recognize it, and send
		// response. if not, send error and list of known commands.
		if _, ok := a.d[bs[0]]; ok {
			a.genMsg(n, reqnum, 101, 1, fmt.Sprintf("dispatching: %v", bs), nil)
			reply, err := a.d[bs[0]](bs[1:])
			if err != nil {
				c.Write([]byte("Sorry, an error occurred and your request could not be completed."))
				a.genMsg(n, reqnum, 500, 3, "failed", err)
			}
			c.Write(reply)
		} else {
			a.genMsg(n, reqnum, 400, 1, fmt.Sprintf("bad command: %v", bs[0]), nil)
			c.Write([]byte(fmt.Sprintf("Unknown command '%v'\nAvailable commands:\n%v",
				bs[0], cmdhelp)))
		}
		// we're done if we're a one-shot connection
		if a.t < 0 {
			a.genMsg(n, reqnum, 198, 2, "closing one-shot", nil)
			return
		}
	}
}
