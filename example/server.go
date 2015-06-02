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
	var socket = flag.String("socket", "localhost:70990", "Addr:port to bind the socket to")
	flag.Parse()

	// set up signal handling to catch SIGINT (^C) and SIGTERM (kill)
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// create the network-facing asock instance
	d := make(asock.Dispatch)
	d["split"] = &asock.DispatchFunc{echosplit, "split"}
	d["nosplit"] = &asock.DispatchFunc{echonosplit, "nosplit"}
	d["time"] = &asock.DispatchFunc{telltime, "nosplit"}

	c := asock.Config{
		Sockname: *socket,
		Msglvl: asock.All,
	}
	as, err := asock.NewTCP(c, d)
	if err != nil {
		panic(err)
	}

	// asock instance is live. serve until an interrupt/kill signal
	// arrives
	log.Println("Asock instance is serving.")
	<-sigchan
	log.Println("OS signal received; shutting down")
	as.Quit()
}
