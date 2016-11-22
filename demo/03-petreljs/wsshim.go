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
	wsaddr = flag.String("wsaddr", "localhost:60607", "Addr:port to bind the websocket to")
	paddr = flag.String("paddr", "localhost:60606", "Addr:port to bind Petrel to")
	connnum = 0
)

func petrelShim(ws *websocket.Conn) {
	connnum++;
	log.Printf("Opening connection %d\n", connnum)
	pc, err := petrel.TCPClient(&petrel.ClientConfig{Addr: *paddr, Timeout: 50})
	if err != nil {
		log.Printf("%d: couldn't instantiate petrel client; bailing: %s\n", connnum, err)
		os.Exit(1)
	}

	for {
		var msg []byte
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Printf("%d: closing conn: couldn't read from ws: %s\n", connnum, err)
			ws.Close()
			pc.Quit()
			break
		}
		resp, err := pc.Dispatch(msg)
		if err != nil {
			log.Printf("%s\n", resp)
			log.Printf("%d: closing conn: couldn't read from petrel: %s\n", connnum, err)
			err = websocket.Message.Send(ws, resp)
			ws.Close()
			pc.Quit()
			break
		}
		log.Printf("%d: dispatched request from ws to petrel\n", connnum)
		err = websocket.Message.Send(ws, resp)
		if err != nil {
			log.Printf("%d: closing conn: couldn't write to websocket: %s\n", connnum, err)
			ws.Close()
			pc.Quit()
			break
		}
		log.Printf("%d: relayed response from petrel to ws", connnum)
	}
}

func main() {
	flag.Parse()
	http.Handle("/shim", websocket.Handler(petrelShim))
	log.Println("wsshim going active. C-c to kill.")
	err := http.ListenAndServe(*wsaddr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
