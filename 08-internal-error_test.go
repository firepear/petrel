package asock

import (
	"fmt"
	"net"
	"testing"
)

// the faulty echo function for our dispatch table
func badecho(s []string) ([]byte, error) {
	return nil, fmt.Errorf("oh no something is wrong")
}

// implement an echo server with a bad command
func TestInternalError(t *testing.T) {
	d := make(Dispatch)    // create Dispatch
	d["echo"] = echo       // and put a function in it
	d["badecho"] = badecho // and a faulty function too
	// instantiate an asocket
	as, err := NewUnix("test08", d, 0, All)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch echoclient
	go echoclient2(as.s, t)
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
	if msg.Txt != "dispatching [badecho foo bar]" {
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
	// shut down asocket
	as.Quit()
}

// this time our (less) fake client will send a string over the
// connection and (hopefully) get it echoed back.
func echoclient2(sn string, t *testing.T) {
	conn, err := net.Dial("unix", sn)
	defer conn.Close()
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	conn.Write([]byte("echo it works!"))
	res, err := readConn(conn)
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "it works!" {
		t.Errorf("Expected 'it works!' but got '%v'", string(res))
	}
	// it's not going to work this time though :(
	conn.Write([]byte("badecho foo bar"))
	res, err = readConn(conn)
	if string(res) != "Sorry, an error occurred and your request could not be completed." {
		t.Errorf("Should have gotten the internal error msg but got '%v'", string(res))
	}
}

