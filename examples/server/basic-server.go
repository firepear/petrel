package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	ps "github.com/firepear/petrel/server"
)

// sigHandler set up OS signal handling for us, returing a channel we
// can watch to know when we've been asked to halt
func sigHandler() chan os.Signal {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	return s
}

// echonosplit is one of the functions we'll use a Handler after we
// instantiate a Server. it echoes its input right back, very
// literally
func echonosplit(args []byte) (uint16, []byte, error) {
	return 200, args, nil
}

// telltime, our other Handler, returns the current datetime
func telltime(args []byte) (uint16, []byte, error) {
	return 200, []byte(time.Now().Format(time.RFC3339)), nil
}

func main() {
	// preliminaries: handle command line args
	var socket = flag.String("socket", "localhost:60606", "Addr:port to bind the socket to")
	var hkey = flag.String("hmac", "", "HMAC secret key")
	flag.Parse()

	// then call a function which sets up OS signal handling for
	// us. this is in signals.go, just to make it clear that the
	// code is unrelated to Petrel
	sigchan := sigHandler()

	// now we're into Petrel related code. create a server
	// configuration which sets Msglvl to "debug" and enables IP
	// logging
	conf := &ps.Config{Addr: *socket, Msglvl: "debug"}
	// and if we've been given an HMAC key, set that
	if *hkey != "" {
		conf.HMACKey = []byte(*hkey)
	}

	// instantiate a server
	s, err := ps.New(conf)
	if err != nil {
		log.Printf("could not instantiate server: %s\n", err)
		os.Exit(1)
	}

	// register our Handler funcs
	if err = s.Register("echo", echonosplit); err != nil {
		log.Printf("failed to register 'echo': %s", err)
		os.Exit(1)
	}
	if err = s.Register("time", telltime); err != nil {
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
		case <-sigchan:
			log.Println("OS sig rec'd")
			s.Quit()
		}
		// there's no default case. that would make it
		// nonblocking, and cause main() to exit immediately.
	}
}
