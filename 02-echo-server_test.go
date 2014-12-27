package adminsock

import (
	"net"
	"testing"
	"time"
)

// implement an echo server
func TestConnHandler(t *testing.T) {
	var d Dispatch   // create Dispatch
	d["echo"] = echo // and put a function in it
	// instantiate an adminsocket
	as, err := New(d, 0)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch fakeclient. we should get a message about the
	// connection.
	go fakeclient(buildSockName(), t)
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "adminsock accepted new connection" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr
	if msg.Err == nil {
		t.Errorf("connection drop should be an err, but got nil")
	}
	if msg.Txt != "adminsock connection lost" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// shut down adminsocket
	as.Quit()
}

// the echo function
func echo(s []string) ([]byte, error) {
	return []byte(s...), nil
}

// this time our fake client will send a single string over the
// connection and (hopefully) get it echoed back.
func fakeclient(sn string, t *testing.T) {
	conn, err := net.Dial("unix", sn)
	defer conn.Close()
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	time.Sleep(100 * time.Millisecond)
}
