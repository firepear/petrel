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
	as, err := New(d, -20707)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// wait 150ms (listener should be killed off in 100)
	time.Sleep(150 * time.Millisecond)
	// check Msgr. It should be ENOLISTENER.
	msg := <-as.Msgr
	if msg.Err == nil {
		t.Errorf("should have gotten an error, but got nil")
	}
	if msg.Txt != "ENOLISTENER" {
		t.Errorf("should have gotten ENOLISTENER, but got: %v", msg.Txt)
	}
	as.Quit()
}

func TestENOLISTENER2(t *testing.T) {
	// implement an echo server
	d := make(Dispatch) // create Dispatch
	d["echo"] = echo    // and put a function in it
	// instantiate an adminsocket
	as, err := New(d, -20707)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// wait
	time.Sleep(100 * time.Millisecond)
	// check Msgr. It should be ENOLISTENER.
	msg := <-as.Msgr
	if msg.Err == nil {
		t.Errorf("should have gotten an error, but got nil")
	}
	if msg.Txt != "ENOLISTENER" {
		t.Errorf("should have gotten ENOLISTENER, but got: %v", msg.Txt)
	}
	// oh no, our adminsocket is dead. gotta spawn a new one.
	as.Quit()
	as, err = New(d, 0)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch echoclient. we should get a message about the
	// connection.
	go echoclient(buildSockName(), t)
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "adminsock conn 1 opened" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr // discard cmd dispatch message
	msg = <-as.Msgr // discard unknown cmd message
	msg = <-as.Msgr
	if msg.Err == nil {
		t.Errorf("connection drop should be an err, but got nil")
	}
	if msg.Txt != "adminsock conn 1 client lost" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// shut down adminsocket
	as.Quit()
}
