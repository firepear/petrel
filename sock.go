package adminsock

// Copyright (c) 2014 Shawn Boyette <shawn@firepear.net>. All rights
// reserved.  Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Socket code for adminsock

import (
	"fmt"
	"net"
	"sync"
	"time"
	
	"firepear.net/goutils/qsplit"
)

// sockAccept monitors the listener socket and spawns connections for
// clients.
func sockAccept(l net.Listener, d Dispatch, t int, m chan *Msg, q chan bool, w *sync.WaitGroup) {
	w.Add(1)
	defer w.Done()
	go sockWatchdog(l, q, w)
	for n := 1; true; n++ {
		c, err := l.Accept()
		if err != nil {
			// is the error because sockWatchdog closed the sock?
			select {
			case <-q:
				// yes; close up shop
				m <- &Msg{"adminsock shutting down", nil}
				return
			default:
				// no, we've had a networking error
				m <- &Msg{"ENOLISTENER", err}
				return
			}
		}
		w.Add(1)
		go connHandler(c, d, n, t, m, w)
	}
}

// sockWatchdog waits to get a signal on the quitter chan, then closes
// it and the listener. This is how we get around sockAccept having a
// blocking event loop.
func sockWatchdog(l net.Listener, q chan bool, w *sync.WaitGroup) {
	defer w.Done()
	<-q        // block until signalled
	q <- true  // relay signal to sockAccept
	l.Close()
}

// connHandler dispatches commands from, and sends reponses to, a client. It
// is launched, per-connection, from sockAccept().
func connHandler(c net.Conn, d Dispatch, n, t int, m chan *Msg, w *sync.WaitGroup) {
	defer w.Done()
	defer c.Close()
	b1 := make([]byte, 64) // buffer 1:  network reads go here, 64B at a time
	var b2 []byte          // buffer 2:  then are accumulated here
	var bs []string        // b2, turned into strings by word
	var cmdhelp string     // list of commands for the auto-help msg
	for cmd, _ := range d {
		cmdhelp = cmdhelp + "    " + cmd + "\n"
	}
	m <- &Msg{fmt.Sprintf("adminsock conn %d opened", n), nil}
	for {
		// set conn timeout deadline if needed
		if t > 0 {
			err := c.SetReadDeadline(time.Now().Add(time.Duration(t) * time.Second))
			if err != nil {
				m <- &Msg{fmt.Sprintf("adminsock conn %d deadline set failed; closing", n), err}
				c.Write([]byte("Sorry, an error occurred. Terminating connection."))
				return
			}
		}
		// get input from the client
		for {
			b, err := c.Read(b1)
			if err != nil {
				m <- &Msg{fmt.Sprintf("adminsock conn %d client lost", n), err}
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
		if _, ok := d[bs[0]]; ok {
			reply, err := d[bs[0]](bs[1:])
			if err != nil {
				c.Write([]byte("Sorry, an error occurred and your request could not be completed."))
				msg := fmt.Sprintf("adminsock conn %d: request failed: %v", n, bs)
				m <- &Msg{msg, err}
			}
			c.Write(reply)
		} else {
			c.Write([]byte(fmt.Sprintf("Unknown command '%v'\nAvailable commands:\n%v",
				bs[0], cmdhelp)))
		}
		// we're done if we're a one-shot connection
		if t < 0 {
			m <- &Msg{fmt.Sprintf("adminsock conn %d closing (one-shot)", n), nil}
			return
		}
	}
}
