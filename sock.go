package adminsock

// Copyright (c) 2014 Shawn Boyette. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Socket code for adminsock

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
)

// launchListener is called by New(). It creates the listener socket
// and termination channel for asAccept(), then launches it as a
// goroutine.
func launchListener() (chan bool, error) {
	// TODO check user. if root, create in /var/run. if not, /tmp. name $0.sock
	l, err := net.Listen("unix", "/tmp/evdadm.sock")
	if err != nil {
		return nil, err
	}
	q := make(chan bool, 1)  // our master off-switch channel
	go asAccept(l, q)
	return q, err
}


// asAccept monitors the listener socket which administrative clients
// connects to, and spawns connections for clients.
func asAccept(l net.Listener, q chan<- bool, r chan<- bool) {
	defer l.Close()        // close it on exit
	switch {
	case conn, err := l.Accept():
		// TODO see conn.SetDeadline for idle timeouts
		if err != nil {
			log.Printf("ERROR Can't make conn on adm sock: %v\n", err)
			l.Close()
			r <- true
			return
		}
		go admHandler(conn, q)
	case <-q:
		break
	}
}


// asHandler dispatches commands from, and talks back to, a client. It
// is launched, per-connection, from asAccept().
func asHandler(c net.Conn, q chan<- bool) {
	log.Println("Accepted connection on adm sock!")
	b1 := make([]byte, 64)  // buffer 1:  network reads go straight here
	b2 := make([]byte, 0)   // buffer 2:  then are accumulated here to handle overruns
	var blen int            // bufferlen: cumulative length of bytes read
	var bstr string         // bufferstr: bytes finally go here when we have them all
	c.Write([]byte(hellomsg))
ReadLoop:
	for {
		defer c.Close()
		for {
			// try to read. n is bytes read.
			n, err := c.Read(b1)
			if err != nil {
				log.Println("Adm socket connection dropped:", err)
				return
			}
			if n > 0 {
				// we read some bytes. first, track how many
				blen += n
				// then copy those bytes into the b2 slice
				b2 = append(b2, b1[:n]...)
				// if we read 64 bytes, loop back to get anything that
				// might be left on the wire
				if n == 64 {
					continue
				}
				// else, we've got a complete command read in. turn it
				// into a string, leaving off the terminal newlines
				bstr = string(b2[:blen-2])
				// reslice b2 so that it will be "empty" on the next read
				b2 = b2[:0]
				// reset total bytes read
				blen = 0
				// break inner loop; drop to switch
				break 
			}
		}
		log.Printf("Read from adm socket: '%s'", bstr)
		switch {
		case bstr == "serverhalt":
			log.Println("Got halt command; shutting down")
			c.Write([]byte("HALTING"))
			q <- true
			return
		case bstr == "help" || bstr == "h" || bstr == "?":
			log.Println("Sending command list")
			if _, err := c.Write([]byte(helpmsg)); err != nil {
				log.Println("Error writing to adm socket; ending connection")
				break ReadLoop
			}
		case bstr == "bye":
			log.Println("Disconnecting adm client")
			if _, err := c.Write([]byte("BYE")); err != nil {
				log.Println("Error writing to adm socket; ending connection")
				break ReadLoop
			}
			return
		default:
			log.Println("Unknown command")
			msg := fmt.Sprintf("Unknown command '%s'. Type 'help' for command list.", bstr)
			if _, err := c.Write([]byte(msg)); err != nil {
				log.Println("Error writing to adm socket; ending connection")
				break ReadLoop
			}
		}
	}
}
