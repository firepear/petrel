package main

import (
	"flag"
	"fmt"
	"strings"

	"firepear.net/asock/client"
)

func main() {
	// handle command line args
	var socket = flag.String("socket", "localhost:60606", "Addr:port to bind the socket to")
	flag.Parse()
	conf := client.Config{Addr: *socket}
	c, err := client.NewTCP(conf)
	if err != nil {
		panic(err)
	}
	req := strings.Join(flag.Args(), " ")
	resp, err := c.Dispatch([]byte(req))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp))
}
