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
				a.sendMsg("adminsock shutting down", nil)
				return
			default:
				// we've had a networking error
				a.Msgr <- &Msg{"ENOLISTENER", err}
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
	var cmdhelp string     // list of commands for the auto-help msg
	for cmd := range a.d {
		cmdhelp = cmdhelp + "    " + cmd + "\n"
	}
	a.sendMsg(fmt.Sprintf("adminsock conn %d opened", n), nil)
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
				a.sendMsg(fmt.Sprintf("adminsock conn %d deadline set failed; closing", n), err)
				c.Write([]byte("Sorry, an error occurred. Terminating connection."))
				return
			}
		}
		// get input from the client
		for {
			b, err := c.Read(b1)
			if err != nil {
				a.sendMsg(fmt.Sprintf("adminsock conn %d client lost", n), err)
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
				bs = qsplit.SplitString(b2)
				// reslice b2 so that it will be "empty" on the next read
				b2 = b2[:0]
				// break inner loop; drop to dispatch
				break 
			}
		}
		// bs[0] is the command. dispatch if we recognize it, and send
		// response. if not, send error and list of known commands.
		if _, ok := a.d[bs[0]]; ok {
			a.sendMsg(fmt.Sprintf("adminsock conn %d dispatching %v", n, bs), nil)
			reply, err := a.d[bs[0]](bs[1:])
			if err != nil {
				c.Write([]byte("Sorry, an error occurred and your request could not be completed."))
				a.sendMsg(fmt.Sprintf("adminsock conn %d: request failed: %v", n, bs), err)
			}
			c.Write(reply)
		} else {
			a.sendMsg(fmt.Sprintf("adminsock conn %d bad cmd: '%v'", n, bs[0]), nil)
			c.Write([]byte(fmt.Sprintf("Unknown command '%v'\nAvailable commands:\n%v",
				bs[0], cmdhelp)))
		}
		// we're done if we're a one-shot connection
		if a.t < 0 {
			a.sendMsg(fmt.Sprintf("adminsock conn %d closing (one-shot)", n), nil)
			return
		}
	}
}

func (a *Adminsock) sendMsg(txt string, err error) {
	select {
	case a.Msgr <- &Msg{txt, err}:
	default:
	}
}
