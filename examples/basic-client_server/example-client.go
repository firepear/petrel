package main

// this is a simple, one-shot client. it connects, sends its
// non-option arguments as a request, prints the result it gets, and
// exits.

import (
	"flag"
	"fmt"
	"os"
	"strings"

	pc "github.com/firepear/petrel/client"
)

func main() {
	// handle command line args
	var socket = flag.String("socket", "localhost:60606", "Where to bind the socket (addr:port)")
	var hkey = flag.String("hmac", "", "HMAC secret key")
	flag.Parse()

	// set up configuration and create client instance
	conf := &pc.ClientConfig{Addr: *socket}
	if *hkey != "" {
		conf.HMACKey = []byte(*hkey)
	}
	c, err := pc.TCPClient(conf)
	if err != nil {
		fmt.Printf("can't initialize client: %s\n", err)
		os.Exit(1)
	}
	defer c.Quit()

	// stitch together the non-option arguments into a request
	req := []byte(strings.Join(flag.Args(), " "))

	// and dispatch it to the server!
	resp, err := c.Dispatch(req)
	if err != nil {
		fmt.Printf("did not get successful response: %s\n", err)
		os.Exit(1)
	}

	// print out what we got back and exit
	fmt.Println(string(resp))
}
