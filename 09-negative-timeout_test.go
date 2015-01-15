package adminsock

import (
	"testing"
)

// sleeperclient is defined in test 07. oneshotclient is defined in
// test 05.

func TestConnNegTimeout(t *testing.T) {
	//
	// rerun timeout test.
	//
	d := make(Dispatch) // create Dispatch
	d["echo"] = echo    // and put a function in it
	// instantiate an adminsocket
	as, err := New("test09", d, -1, All)
	t.Log(as.t)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch sleeperclient. we should get a message about the
	// connection.
	go sleeperclient(as.s, t)
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "adminsock conn 1 opened" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr
	if msg.Err == nil {
		t.Errorf("connection drop should be an err, but got nil")
	}
	if msg.Txt != "adminsock conn 1 client lost" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	//
	// now rerun oneshot test
	//
	go oneshotclient(as.s, t)
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "adminsock conn 2 opened" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection drop should be nil, but got %v", err)
	}
	if msg.Txt != "adminsock conn 2 closing (one-shot)" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}	
	// shut down adminsocket
	as.Quit()
}
