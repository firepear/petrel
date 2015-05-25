package asock

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

// functions echo() and readConn() are defined in test 02. multiclient
// is defined in test 03.

func TestMultiServer2(t *testing.T) {
	// implement an echo server
	d := make(Dispatch) // create Dispatch
	d["echo"] = &DispatchFunc{echo, "split"} // and put a function in it
	// instantiate an asocket
	c := Config{"/tmp/test04.sock", 0, 32, All, nil}
	as, err := NewUnix(c, d)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch clients
	rand.Seed(time.Now().Unix())
	x := 3
	for i := 0; i < x; i++ {
		go multiclient(as.s, t)
	}
	// wait for all clients to connect
	for i := 0; i < x; i++ {
		for {
			msg := <-as.Msgr
			if strings.Contains(msg.Txt, "connected") {
				break
			}
		}
	}
	// do not wait for disconnect Msg. rely on shutdown to handle
	// things appropriately. This is actually the test in this file.
	as.Quit()	
}

