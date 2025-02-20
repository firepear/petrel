package main

// this is a simple, one-shot client. it connects, sends its
// non-option arguments as a request, prints the result it gets, and
// exits.

import (
	"flag"
	"fmt"

	pc "github.com/firepear/petrel/client"
)

func main() {
	// handle command line args
	var socket = flag.String("socket", "localhost:60606", "Where to bind the socket (addr:port)")
	var hkey = flag.String("hmac", "", "HMAC secret key")
	flag.Parse()

	// first argument is our request
	if len( flag.Args() ) == 0 {
		fmt.Printf("usage: go run example-client.go [REQUEST] [PAYLOAD]\n")
		return
	}
	req := flag.Args()[0]

	// stitch together the non-option arguments into a payload
	payload := []byte("foo") //strings.Join(flag.Args()[1:], " "))

	// set up configuration and create client instance
	conf := &pc.Config{Addr: *socket}
	if *hkey != "" {
		conf.HMACKey = []byte(*hkey)
	}
	c, err := pc.New(conf)
	if err != nil {
		fmt.Printf("can't initialize client: %s\n", err)
		return
	}
	defer c.Quit()

	// and dispatch request and payload to the server!
	err = c.Dispatch(req, payload)
	if err != nil {
		fmt.Printf("did not get successful response: %s\n", err)
		fmt.Printf("req: %s, status: %d, payload: %v\n",
			c.Resp.Req, c.Resp.Status, c.Resp.Payload)
		return
	}
	if c.Resp.Status != 200 {
		fmt.Printf("req '%s' status %d: %s\n",
			c.Resp.Req, c.Resp.Status, c.StatusTxt())
		return
	}

	// print out what we got back and exit
	fmt.Println(string(c.Resp.Payload))
}
