package petrel

import (
	"strings"
	"testing"
)

// implement an echo server
func TestServReqlen(t *testing.T) {
	// instantiate petrel
	c := &ServerConfig{Sockname: "/tmp/test05c.sock", Msglvl: All, Reqlen: 10}
	as, err := UnixServ(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.AddFunc("echo", "args", echo)

	// launch a client and do some things
	go reqclient("/tmp/test05c.sock", t)
	reqtests(as, t)
	// shut down petrel
	as.Quit()
}

func reqtests(as *Server, t *testing.T) {
	// we should get a message about the connection.
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if !strings.HasPrefix(msg.Txt, "client connected") {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// and a message about dispatching the command
	msg = <-as.Msgr
	if msg.Txt != "request over limit; closing conn" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 402 {
		t.Errorf("msg.Code should have been 402 but got: %v", msg.Code)
	}
}

// this time our (less) fake client will send a string over the
// connection and (hopefully) get it echoed back.
func reqclient(sn string, t *testing.T) {
	ac, err := UnixClient(&ClientConfig{Addr: sn})
	if err != nil {
		t.Fatalf("client instantiation failed! %s", err)
	}
	defer ac.Close()

	resp, err := ac.Dispatch([]byte("echo this string is way too long! it won't work!"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "PERRPERR402" {
		t.Errorf("Expected 'PERRPERR402' but got '%v'", string(resp))
	}
}
