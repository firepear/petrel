// wsshim implements a standalone websocket-to-petrel bridge or shim.
package main

import (
	"flag"
	"log"
	"os"
	"net/http"

	"golang.org/x/net/websocket"
	"firepear.net/petrel"
)

var (
	wsaddr = flag.String("wsaddr", "localhost:60606", "Addr:port to bind the websocket to")
	paddr = flag.String("paddr", "localhost:60607", "Addr:port to bind Petrel to")
	connnum = 0
)

func petrelShim(ws *websocket.Conn) {
	connnum++;
	log.Printf("connection %d opened\n", connnum)
	var b1 = make([]byte, 128)
	var b2 = []byte{}

	pc, err := petrel.TCPClient(&petrel.ClientConfig{Addr: *paddr, Timeout: 50})
	if err != nil {
		log.Printf("%d: couldn't instantiate petrel client; bailing: %s", connnum, err)
		os.Exit(1)
	}

WSLoop:
	for {
		var n int
		b2 = b2[:]
		for {
			n, err := ws.Read(b1)
			if err != nil {
				log.Printf("%d: couldn't read from websocket; closing conn: %s", connnum, err)
				ws.Close();
				break WSLoop
			}
			if n < 128 {
				break
			}
			b2 = append(b2, b1...)
		}
		b2 = append(b2, b1[:n]...)
	}
}

func main() {
	flag.Parse()
	http.Handle("/shim", websocket.Handler(petrelShim))
	err := http.ListenAndServe(*wsaddr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
