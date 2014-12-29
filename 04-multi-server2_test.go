package adminsock

import (
	"math/rand"
	"testing"
	"time"
)

// functions echo() and readConn() are defined in test 02. multiclient
// is defined in test 03.

func TestMultiServer2(t *testing.T) {
	// implement an echo server
	d := make(Dispatch) // create Dispatch
	d["echo"] = echo    // and put a function in it
	// instantiate an adminsocket
	as, err := New(d, 0)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch clients
	rand.Seed(time.Now().Unix())
	x := 3
	for i := 0; i < x; i++ {
		go multiclient(buildSockName(), t)
	}
	for i := 0; i < x; i++ {
		msg := <-as.Msgr
		if msg.Err != nil {
			t.Errorf("connection creation returned error: %v", msg.Err)
		}
	}
	// do not wait for disconnect Msg. rely on shutdown to handle
	// things appropriately. This is actually the test in this file.
	as.Quit()	
}
