package main

// This is a simple, one-shot client. It connects, sends its
// non-option arguments as a request, prints the result it gets, and
// exits.

import (
	"flag"
	"fmt"
	"strings"

	"firepear.net/petrel"
)

func main() {
	// handle command line args
	var socket = flag.String("socket", "localhost:60606", "Addr:port to bind the socket to")
	flag.Parse()

	// set up configuration and create client instance
	conf := &petrel.ClientConfig{Addr: *socket}
	c, err := petrel.TCPClient(conf)
	if err != nil {
		panic(err)
	}
	defer c.Quit()

	// stitch together the non-option arguments into our request
	req := strings.Join(flag.Args(), " ")

	// and dispatch it to the server!
	resp, err := c.Dispatch([]byte(req))
	if err != nil {
		panic(err)
	}

	// print out what we got back and exit
	fmt.Println(string(resp))
}
