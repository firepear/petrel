package adminsock

import (
	"testing"
	"time"
)

// functions echo() and readConn() are defined in test 02. multiclient
// is defined in test 03.

func TestENOLISTENER(t *testing.T) {
	// implement an echo server
	d := make(Dispatch) // create Dispatch
	d["echo"] = echo    // and put a function in it
	// instantiate an adminsocket
	as, err := New(d, -42)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// wait
	time.Sleep(100 * time.Millisecond)
	// check Msgr until an error comes up. It should be ENOLISTENER.
	msg := <-as.Msgr
	if msg.Err == nil {
		t.Errorf("should have gotten an error, but got nil")
	}
	if msg.Txt != "ENOLISTENER" {
		t.Errorf("should have gotten ENOLISTENER, but got: %v", msg.Txt)
	}
}
