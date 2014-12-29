package adminsock

// Copyright (c) 2014 Shawn Boyette <shawn@firepear.net>. All rights
// reserved.  Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Socket code for adminsock

import (
	"fmt"
	"net"
	"sync"

	"firepear.net/goutils/qsplit"
)

// sockAccept monitors the listener socket and spawns connections for
// clients.
func sockAccept(l net.Listener, d Dispatch, t int, m chan *Msg, q chan bool, w *sync.WaitGroup) {
	defer w.Done()
	w.Add(1)
	go sockWatchdog(l, q, w)
	// TODO make list of known commands and hand them to connHandlers
	// for better "unknown command" handling
	for {
		// TODO see conn.SetDeadline for idle timeouts
		c, err := l.Accept()
		if err != nil {
			// is the error because sockWatchdog closed the sock?
			select {
			case <-q:
				// yes; close up shop
				m <- &Msg{"adminsock shutting down", nil}
				close(m)
				return
			default:
				// no, we've had a networking error
				m <- &Msg{"ENOSOCK" ,err}
				close(m)
				q <- true // kill off the watchdog
				return
			}
		}
		w.Add(1)
		go connHandler(c, d, m, w)
	}
}

// sockWatchdog waits to get a signal on the quitter chan, then closes
// it and the listener.
func sockWatchdog(l net.Listener, q chan bool, w *sync.WaitGroup) {
	defer w.Done()
	<-q        // block until signalled
	l.Close()
	q <- true  // signal to sockAccept
	close(q)
}

// connHandler dispatches commands from, and talks back to, a client. It
// is launched, per-connection, from sockAccept().
func connHandler(c net.Conn, d Dispatch, m chan *Msg, w *sync.WaitGroup) {
	defer w.Done()
	defer c.Close()
	b1 := make([]byte, 64) // buffer 1:  network reads go here, 64B at a time
	var b2 []byte          // buffer 2:  then are accumulated here
	var bs []string        // b2, turned into strings by word
	var cmdhelp string
	for cmd, _ := range d {
		cmdhelp = cmdhelp + "    " + cmd + "\n"
	}
	m <- &Msg{"adminsock accepted new connection", nil}
	// 
	for {
		for {
			// try to read. n is bytes read.
			n, err := c.Read(b1)
			if err != nil {
				m <- &Msg{"adminsock connection dropped", err}
				return
			}
			if n > 0 {
				// then copy those bytes into the b2 slice
				b2 = append(b2, b1[:n]...)
				// if we read 64 bytes, loop back to get anything that
				// might be left on the wire
				if n == 64 {
					continue
				}
				bs = qsplit.SplitString(b2)
				// reslice b2 so that it will be "empty" on the next read
				b2 = b2[:0]
				// break inner loop; drop to dispatch
				break 
			}
		}
		if _, ok := d[bs[0]]; ok {
			// dispatch command if we know about it
			reply, err := d[bs[0]](bs[1:])
			if err != nil {
				c.Write([]byte("Sorry, an error occurred and your request could not be completed."))
				msg := fmt.Sprintf("adminsock error: request failed: '%v'", bs)
				m <- &Msg{msg, err}
			}
			c.Write(reply)
		} else {
			c.Write([]byte(fmt.Sprintf("Unknown command %v\nAvailable commands: %v\n",
				bs[0], cmdhelp)))
		}
	}
}
