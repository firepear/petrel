package adminsock

import (
	"fmt"
	"net"
	"testing"
)

// the faulty echo function for our dispatch table
func badecho(s []string) ([]byte, error) {
	return nil, fmt.Errorf("oh no something is wrong!")
}

// implement an echo server with a bad command
func TestInternalError(t *testing.T) {
	d := make(Dispatch)    // create Dispatch
	d["echo"] = echo       // and put a function in it
	d["badecho"] = badecho // and a faulty function too
	// instantiate an adminsocket
	as, err := New(d, 0)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch fakeclient. we should get a message about the
	// connection.
	go echoclient2(buildSockName(), t)
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
	if msg.Txt != "adminsock conn 1: request failed: [badecho foo bar]" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// shut down adminsocket
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
