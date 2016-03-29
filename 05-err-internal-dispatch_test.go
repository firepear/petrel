package petrel

import (
	"fmt"
	"testing"

	"firepear.net/pclient"
)

// the faulty echo function for our dispatch table
func badecho(s [][]byte) ([]byte, error) {
	return nil, fmt.Errorf("oh no something is wrong")
}

// implement an echo server with a bad command
func TestInternalError(t *testing.T) {
	// instantiate petrel
	c := &Config{Sockname: "/tmp/test08.sock", Msglvl: All}
	as, err := NewUnix(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.AddHandlerFunc("echo", "args", echo)
	as.AddHandlerFunc("badecho", "args", badecho)

	// launch echoclient
	go internalerrclient(as.s, t)
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "client connected" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	msg = <-as.Msgr // ignore msg on dispatch
	msg = <-as.Msgr // ignore msg of reply sent
	// wait for msg from unsuccessful command
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("unsuccessful cmd shouldn't be err, but got %v", err)
	}
	if msg.Txt != "dispatching [badecho]" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	msg = <-as.Msgr
	if msg.Txt != "request failed" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Err.Error() != "oh no something is wrong" {
		t.Errorf("unsuccessful cmd should be an error, but got %v", msg.Err)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr
	if msg.Err == nil {
		t.Errorf("connection drop should be an err, but got nil")
	}
	if msg.Txt != "client disconnected" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// shut down petrel
	as.Quit()
}

// this time our (less) fake client will send a string over the
// connection and (hopefully) get it echoed back.
func internalerrclient(sn string, t *testing.T) {
	ac, err := pclient.NewUnix(&pclient.Config{Addr: sn})
	if err != nil {
		t.Fatalf("pclient instantiation failed! %s", err)
	}
	defer ac.Close()

	resp, err := ac.Dispatch([]byte("echo it works!"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "it works!" {
		t.Errorf("Expected 'it works!' but got '%v'", string(resp))
	}
	// it's not going to work this time though :(
	resp, err = ac.Dispatch([]byte("badecho foo bar"))
	if string(resp) != "Sorry, an error occurred and your request could not be completed." {
		t.Errorf("Should have gotten the internal error msg but got '%v'", string(resp))
	}
}

