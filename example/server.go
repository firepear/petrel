package main

import (
	//"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"firepear.net/asock"
)

///////////////////////////////////////////////// dispatch functions

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

//func hellojson(args [][]byte) ([]byte, error) {
//}

////////////////////////////////////////////////////////////////////

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
	msg := <-msgchan
	log.Println("Asock instance has shut down. Last Msg received was:")
	log.Println(msg)
}

func msgHandler(as *asock.Asock, sigchan chan os.Signal, msgchan chan error) {
	var msg *asock.Msg
	keepgoing := true
Loop:
	for keepgoing {
		select {
		case msg = <-as.Msgr:
			switch msg.Code {
			case 599:
				// 599 is "the asock listener has died". this means
				// we're not accepting connections anymore. shutdown
				// our Asock and break out of the 'for' so our program
				// knows what happened.
				as.Quit()
				keepgoing = false
				msgchan <- msg
				break Loop
			case 199:
				// 199 is "we've been told to quit", so break out of
				// the 'for'.
				keepgoing = false
				msgchan <- msg
				break Loop
			default:
				log.Println(msg)
			}
		case <- sigchan:
			// we've trapped a signal from the OS. tell our Asock to
			// shut down. a subsequent pass through this loop should
			// be a Msg with Code = 199, which will terminate the loop
			// adn return from this function.
			log.Println("OS signal received; shutting down")
			as.Quit()
		}
	}
}
