package main

import (
	"flag"
	"log"
	"os"
	"time"

	ps "github.com/firepear/petrel/server"
)

// echonosplit is one of the functions we'll use as Responders after
// we instantiate a Server. it's an echo function, with an argmode of
// "blob".
func echonosplit(args []byte) ([]byte, error) {
	return args, nil
}

// telltime, our other Responder, returns the current datetime
func telltime(args []byte) ([]byte, error) {
	return []byte(time.Now().Format(time.RFC3339)), nil
}

///////////////////////////////////////////////////////////////////////////

// msgHandler is a function which we'll launch later on as a
// goroutine. It listens to our Server's Msgr channel, checking for a
// few critical things and logging everything else informationally.
func msgHandler(s *ps.Server, msgchan chan error) {
	keepalive := true

	for keepalive {
		// wait on a Msg to arrive and do a switch based on
		// its status code.
		msg := <-s.Msgr
		switch msg.Code {
		case 599:
			// 599 is "the Server listener socket has
			// died". this means we're not accepting
			// connections anymore. call s.Quit() to clean
			// things up, send the Msg to our main
			// routine, then kill this loop
			s.Quit()
			keepalive = false
			msgchan <- msg
		case 199:
			// 199 is "we've been told to quit", so we
			// want to break out of the loop here as well
			keepalive = false
			msgchan <- msg
		default:
			// anything else we'll log to the console to
			// show what's going on under the hood!
			log.Println(msg)
		}
	}
}

///////////////////////////////////////////////////////////////////////////

func main() {
	// first, handle command line args
	var socket = flag.String("socket", "localhost:60606", "Addr:port to bind the socket to")
	var hkey = flag.String("hmac", "", "HMAC secret key")
	flag.Parse()

	// set up our Petrel instance.  first create a configuration
	c := &ps.Config{Sockname: *socket, Msglvl: "debug"}
	if *hkey != "" {
		c.HMACKey = []byte(*hkey)
	}

	// then instantiate a Server.
	s, err := ps.TCPServer(c)
	if err != nil {
		log.Printf("could not instantiate Server: %s\n", err)
		os.Exit(1)
	}

	// then Register our Responders with the Server
	err = s.Register("echo", echonosplit)
	if err != nil {
		log.Printf("failed to register responder 'echo': %s", err)
		os.Exit(1)
	}
	err = s.Register("time", telltime)
	if err != nil {
		log.Printf("failed to register responder 'time': %s", err)
		os.Exit(1)
	}
	// now, if a client sends a request starting with "echo", the
	// request will be dispatched to echonosplit. likewise "time"
	// and telltime.
	log.Println("Petrel handler is serving.")

	// the Server is now listening and ready to do work.  it's
	// time to start msgHandler, the event loop we defined
	// earlier. we hand it a channel that it uses to pass
	// important Msgs to the main event loop, which is coming up
	// next.  it's a 'chan error' instead of a 'chan petrel.Msg'
	// because petrel.Msg implements the error interface.
	msgchan := make(chan error, 1)
	go msgHandler(s, msgchan)

	// and here is the main eventloop. it's simply a select on
	// msgchan and sigchan, so that we can handle shutdown
	// cleanly. the work of handling requests is entirely inside
	// Petrel, and requires no application logic or intervention.
	keepalive := true
	for keepalive {
		select {
		case msg := <-msgchan:
			// we've been handed a Msg over msgchan, which
			// means that our Server has shut itself down
			// for some reason. we're going to exit this
			// loop, causing main() to terminate.
			log.Printf("Handler has shut down. Last Msg received was: %s", msg)
			keepalive = false
			break
		case <-s.Sig:
			// we've trapped a signal from the OS. tell
			// our Server to shut down, but don't exit the
			// eventloop because we want to handle the
			// Msgs which will be incoming -- including
			// the one we'll get on msgchan once the
			// Server has finished its work.
			log.Println("OS signal received; shutting down")
			s.Quit()
		}
		// there's no default case in the select, as that
		// would make it nonblocking, which would in turn
		// cause main() to exit immediately.
	}
}
