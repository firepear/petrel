package petrel

import (
	"testing"

	"firepear.net/pclient"
)

// implement an echo server
func TestReqlen(t *testing.T) {
	// instantiate petrel
	c := &Config{Sockname: "/tmp/test05c.sock", Msglvl: All, Reqlen: 10}
	as, err := NewUnix(c, 700)
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

func reqtests(as *Handler, t *testing.T) {
	// we should get a message about the connection.
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
		t.Errorf("successful cmd shouldn't be err, but got %v", msg.Err)
	}
	if msg.Txt != "request over limit; closing conn" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 502 {
		t.Errorf("msg.Code should have been 502 but got: %v", msg.Code)
	}
}

// this time our (less) fake client will send a string over the
// connection and (hopefully) get it echoed back.
func reqclient(sn string, t *testing.T) {
	ac, err := pclient.NewUnix(&pclient.Config{Addr: sn})
	if err != nil {
		t.Fatalf("pclient instantiation failed! %s", err)
	}
	defer ac.Close()

	resp, err := ac.Dispatch([]byte("echo this string is way too long! it won't work!"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "PERRPERR402 Request over limit" {
		t.Errorf("Expected 'PERRPERR402 Request over limit' but got '%v'", string(resp))
	}
}
