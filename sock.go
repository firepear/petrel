package adminsock

// Copyright (c) 2014 Shawn Boyette. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Socket code for adminsock

import (
	"net"
	"os"
	"os/user"
)

// launchListener is called by New(). It creates the listener socket
// and termination channel for asAccept(), then launches it as a
// goroutine.
func launchListener() (chan bool, chan error, error) {
	// TODO accept map of functions to be passed
	var l net.Listener
	u, err := user.Current()
	if err != nil {
		return nil, nil, err
	}
	if u.Uid == "0" {
		l, err = net.Listen("unix", "/var/run/" + os.Args[0] + ".sock")
	} else {
		l, err = net.Listen("unix", "/tmp/" + os.Args[0] + ".sock")
	}
	if err != nil {
		return nil, nil, err
	}
	q := make(chan bool, 1)   // master off-switch channel
	e := make(chan error, 32) // error reporting
	go sockAccept(l, q, e)
	return q, e, err
}

// sockAccept monitors the listener socket which administrative clients
// connects to, and spawns connections for clients.
func sockAccept(l net.Listener, q chan bool, e chan error) {
	defer l.Close()        // close it on exit
	for {
		// TODO see conn.SetDeadline for idle timeouts
		conn, err := l.Accept()
		if err != nil {
			// TODO we can't talk to teh listener anymore. send a
			// fatal on the error channel to let our user know we're
			// shutting down and they should call New() again
			l.Close()
			return
		}
		go connHandler(conn, q, e)
	}
}


// connHandler dispatches commands from, and talks back to, a client. It
// is launched, per-connection, from sockAccept().
func connHandler(c net.Conn, q chan bool, e chan error) {
	defer c.Close()
	//log.Println("Accepted connection on adm sock!")
	b1 := make([]byte, 64)  // buffer 1:  network reads go straight here
	b2 := make([]byte, 0)   // buffer 2:  then are accumulated here to handle overruns
	var blen int            // bufferlen: cumulative length of bytes read
	var bstr string         // bufferstr: bytes finally go here when we have them all
	//c.Write([]byte(hellomsg))
//ReadLoop:
	for {
		for {
			// try to read. n is bytes read.
			n, err := c.Read(b1)
			if err != nil {
				//log.Println("Adm socket connection dropped:", err)
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
				// into a string
				bstr = string(b2)
				// reslice b2 so that it will be "empty" on the next read
				b2 = b2[:0]
				// reset total bytes read
				blen = 0
				// break inner loop; drop to dispatch
				break 
			}
		}
		// TODO bstr dispatch table action goes here. fake it for now
		// to get around compile errors
		c.Write([]byte(bstr))

		//switch {
		//default:
		//	log.Println("Unknown command")
		//	msg := fmt.Sprintf("Unknown command '%s'. Type 'help' for command list.", bstr)
		//	if _, err := c.Write([]byte(msg)); err != nil {
		//		log.Println("Error writing to adm socket; ending connection")
		//		break ReadLoop
		//	}
		//}
	}
}
