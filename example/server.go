package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"firepear.net/asock"
)

func main() {
	// handle command line args
	var socket = flag.String("socket", "localhost:60606", "Addr:port to bind the socket to")
	flag.Parse()

	// set up signal handling to catch SIGINT (^C) and SIGTERM (kill)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// create the network-facing asock instance. First set up the
	// Dispatch struct.
	d := make(asock.Dispatch)
	d["split"] = &asock.DispatchFunc{echosplit, "split"}
	d["nosplit"] = &asock.DispatchFunc{echonosplit, "nosplit"}
	d["time"] = &asock.DispatchFunc{telltime, "nosplit"}
	// then the Asock configuration
	c := asock.Config{
		Sockname: *socket,
		Msglvl: asock.All,
	}
	// then do the instantiation
	as, err := asock.NewTCP(c, d)
	if err != nil {
		panic(err)
	}
	log.Println("Asock instance is serving.")

	// no error, so the asock instance is live. spin up our Msgr handler.
	msgchan := make(chan error, 1)
	go msgHandler(as, sigchan, msgchan)

	// this is our eventloop. since we're a bare-bones example server,
	// we just do a select on msgchan and sigchan, waiting on a
	// notification via one of those channels.  as part of a "real"
	// application, there would be more things to handle here.
	keepalive := true
	for keepalive {
		select {
		case msg := <-msgchan:
			// we've been handed a Msg over msgchan, which means that
			// our Asock has shut itself down for some reason. if this
			// were a more robust server, we would modularize Asock
			// creation and this eventloop, so that should we trap a
			// 599 we could spawn a new Asock and launch it in this
			// one's place. but we're just gonna exit this loop,
			// causing main() to terminate, and with it the server
			// instance.
			log.Println("Asock instance has shut down. Last Msg received was:")
			log.Println(msg)
			keepalive = false
			break
		case <- sigchan:
			// we've trapped a signal from the OS. tell our Asock to
			// shut down, but don't exit the eventloop because we want
			// to handle the Msgs which will be incoming.
			log.Println("OS signal received; shutting down")
			as.Quit()
		}
	}
}

func msgHandler(as *asock.Asock, sigchan chan os.Signal, msgchan chan error) {
	var msg *asock.Msg
	keepalive := true
	for keepalive {
		msg = <-as.Msgr
		switch msg.Code {
		case 599:
			// 599 is "the asock listener has died". this means we're
			// not accepting connections anymore. shutdown our Asock
			// and break out of the 'for' so our program knows what
			// happened.
			as.Quit()
			keepalive = false
			msgchan <- msg
		case 199:
			// 199 is "we've been told to quit", so break out of the
			// 'for'.
			keepalive = false
			msgchan <- msg
		default:
			log.Println(msg)
		}
	}
}

func echosplit(args [][]byte) ([]byte, error) {
	var b []byte
	for _, arg := range args {
		b = append(b, arg...)
	}
	return b, nil
}

func echonosplit(args [][]byte) ([]byte, error) {
	return args[0], nil
}

func telltime(args [][]byte) ([]byte, error) {
	return []byte(time.Now().Format(time.RFC3339)), nil
}
