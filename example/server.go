package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"firepear.net/asock"
)

func main() {
	// first, handle command line args
	var socket = flag.String("socket", "localhost:60606", "Addr:port to bind the socket to")
	flag.Parse()

	// now let's give ourselves a way to shut down. we'll listen for
	// SIGINT and SIGTERM, so we can behave like a proper service
	// (mostly -- we're not writing out a pidfile). anyway, to do that
	// we need a channel to recieve signal notifications on.
	sigchan := make(chan os.Signal, 1)
	// and then we register sigchan to listen for the signals we want.
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// with that done, we can set up our Asock instance.  first we set
	// up the Asock configuration
	c := asock.Config{
		Sockname: *socket,
		Msglvl: asock.All,
	}
	// and then we call the constructor! in a real server, obviously,
	// you wouldn't want to just panic.
	as, err := asock.NewTCP(c)
	if err != nil {
		panic(err)
	}
	// and add our command handlers, using a map for convenient error
	// checking
	handlers := map[string]func([][]byte) ([]byte, error){
		"split": echosplit,
		"nosplit": echonosplit,
		"time": telltime,
		"badcmd": thisfuncerrs,
	}
	for name, function := range handlers {
		err = as.AddHandler(name, "split", function)
		if err != nil {
			panic(err)
		}
	}
	log.Println("Asock instance is serving.")

	// at this point, our Asock (as) is listening and ready to do its
	// thing. it's time to spin up an event loop which listens to
	// as.Msgr, which is how as tells us what it's doing. first we
	// need a channel so that we can get *some* messages out of that
	// loop. why is it a 'chan error' instead of a 'chan asock.Msg'?
	// asock.Msg implements error, so we *can*, basically. less
	// typing.
	msgchan := make(chan error, 1)
	// now launch the handler as a goroutine.
	go msgHandler(as, msgchan)

	// this is our *main* eventloop. since we're a bare-bones example
	// server, we just do a select on msgchan and sigchan, waiting on
	// a notification via one of those channels.  as part of a "real"
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
		// there's no default case in the select, as that would cause
		// it to be nonblocking. and that would cause main() to exit
		// immediately.
	}
}

func msgHandler(as *asock.Asock, msgchan chan error) {
	// our Msg handler function is very simple. it's almost a clone of
	// the main eventloop. first we just create a couple of variables
	// to hold Msgs and to control the for loop.
	var msg *asock.Msg
	keepalive := true

	for keepalive {
		// then we wait on a Msg to arrive and do a switch based on
		// its status code.
		msg = <-as.Msgr
		switch msg.Code {
		case 599:
			// 599 is "the Asock listener has died". this means we're
			// not accepting connections anymore. call as.Quit() to
			// clean things up, send the Msg to our main routine, then
			// kill this for loop
			as.Quit()
			keepalive = false
			msgchan <- msg
		case 199:
			// 199 is "we've been told to quit", so we want to break
			// out of the 'for' here as well
			keepalive = false
			msgchan <- msg
		default:
			// anything else we just log!
			log.Println(msg)
		}
	}
}

// this handler is an echo function, with an argmode of "split".
func echosplit(args [][]byte) ([]byte, error) {
	var b []byte
	for _, arg := range args {
		b = append(b, arg...)
		b = append(b, 32)
	}
	return b, nil
}

// this handler is an echo function, with an argmode of
// "nosplit".
func echonosplit(args [][]byte) ([]byte, error) {
	return args[0], nil
}

// this one just returns the current datetime
func telltime(args [][]byte) ([]byte, error) {
	return []byte(time.Now().Format(time.RFC3339)), nil
}

// and this one returns an error
func thisfuncerrs(args [][]byte) ([]byte, error) {
	return nil, errors.New("Something went wrong inside me!")
}
