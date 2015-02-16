package asock

import (
	"testing"
)

// the echo function for our dispatch table
func echonosplit(args [][]byte) ([]byte, error) {
	return args[0], nil
}

// implement an echo server
func TestEchoNosplit(t *testing.T) {
	d := make(Dispatch) // create Dispatch
	d["echo"] = echonosplit    // and put a function in it
	// instantiate an asocket
	c := Config{"/tmp/test12.sock", 0, "nosplit", All}
	as, err := NewUnix(c, d)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch echoclient. we should get a message about the
	// connection.
	go echoclient(as.s, t)
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "client connected" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// and a message about dispatching the command
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("successful cmd shouldn't be err, but got %v", err)
	}
	if msg.Txt != "dispatching [echo]" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 101 {
		t.Errorf("msg.Code should have been 101 but got: %v", msg.Code)
	}
	// and a message that we have replied
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("successful cmd shouldn't be err, but got %v", err)
	}
	if msg.Txt != "reply sent" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 200 {
		t.Errorf("msg.Code should have been 200 but got: %v", msg.Code)
	}
	// wait for msg from unsuccessful command
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("unsuccessful cmd shouldn't be err, but got %v", err)
	}
	if msg.Txt != "bad command 'foo'" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 400 {
		t.Errorf("msg.Code should have been 400 but got: %v", msg.Code)
	}
	// wait for disconnect Msg
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
