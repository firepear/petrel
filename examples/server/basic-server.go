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
func echonosplit(args []byte) (uint16, []byte, error) {
	return 200, args, nil
}

// telltime, our other Responder, returns the current datetime
func telltime(args []byte) (uint16, []byte, error) {
	return 200, []byte(time.Now().Format(time.RFC3339)), nil
}

func main() {
	// handle command line args
	var socket = flag.String("socket", "localhost:60606", "Addr:port to bind the socket to")
	var hkey = flag.String("hmac", "", "HMAC secret key")
	flag.Parse()

	// create a basic server configuration
	conf := &ps.Config{Addr: *socket, Msglvl: "debug", LogIP: true}
	// and if we've been given an HMAC key, set that
	if *hkey != "" {
		conf.HMACKey = []byte(*hkey)
	}

	// instantiate a Server.
	s, err := ps.New(conf)
	if err != nil {
		log.Printf("could not instantiate server: %s\n", err)
		os.Exit(1)
	}
	// then strip out its PROTOCHECK handler
	//ok := s.RemoveHandler("PROTOCHECK")
	//if !ok {
	//	log.Println("removing PROTOCHECK failed")
	//	os.Exit(1)
	//}

	// Register our Handler funcs
	err = s.Register("echo", echonosplit)
	if err != nil {
		log.Printf("failed to register 'echo': %s", err)
		os.Exit(1)
	}
	err = s.Register("time", telltime)
	if err != nil {
		log.Printf("failed to register 'time': %s", err)
		os.Exit(1)
	}
	// now, if a client sends a request starting with "echo", the
	// request will be dispatched to echonosplit. likewise "time"
	// and telltime.

	// the Server is now listening and ready to do work, so we
	// start an event loop. it's simply a select on theh server's
	// Shutdown channel so that we can handle that event
	// cleanly. the work of handling requests is entirely inside
	// Petrel, and requires no application logic or intervention.
	keepalive := true
	for keepalive {
		select {
		case msg := <-s.Shutdown:
			// we've been handed a Msg here, which means
			// that our Server has shut itself down for
			// some reason. set keepalive to false and
			// break to exit the select, which then causes
			// the 'for' to end and main() to terminate.
			log.Println("app event loop:", msg)
			keepalive = false
			break
		}
		// there's no default case. that would make it
		// nonblocking, and cause main() to exit immediately.
	}
}
