package asock

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
	// instantiate an asocket
	as, err := NewUnix("test06-1", d, -20707, All)
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
	if msg.Txt != "read from listener socket failed" {
		t.Errorf("unexpected message: %v", msg.Txt)
	}
	if msg.Code != 599 {
		t.Errorf("dead listener should be code 599 but got %v", msg.Code)
	}
	as.Quit()
}

func TestENOLISTENER2(t *testing.T) {
	// implement an echo server
	d := make(Dispatch) // create Dispatch
	d["echo"] = echo    // and put a function in it
	// instantiate an asocket
	as, err := NewUnix("test06-2", d, -20707, All)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// wait
	time.Sleep(150 * time.Millisecond)
	// check Msgr. It should be ENOLISTENER.
	msg := <-as.Msgr
	if msg.Err == nil {
		t.Errorf("should have gotten an error, but got nil")
	}
	if msg.Txt != "read from listener socket failed" {
		t.Errorf("unexpected message: %v", msg.Txt)
	}
	// oh no, our asocket is dead. gotta spawn a new one.
	as.Quit()
	as, err = NewUnix("test06-3", d, 0, All)
	if err != nil {
		t.Errorf("Couldn't spawn second listener: %v", err)
	}
	// launch echoclient. we should get a message about the
	// connection.
	go echoclient(as.s, t)
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "client connected" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr // discard cmd dispatch message
	msg = <-as.Msgr // discard reply sent message
	msg = <-as.Msgr // discard unknown cmd message
	msg = <-as.Msgr
	if msg.Err == nil {
		t.Errorf("connection drop should be an err, but got nil")
	}
	if msg.Txt != "client disconnected" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// shut down asocket
	as.Quit()
}
