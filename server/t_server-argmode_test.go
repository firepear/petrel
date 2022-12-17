package server

import (
	"strings"
	"testing"

	pc "github.com/firepear/petrel/client"
)

// the echo function for our dispatch table
func echonosplit(args []byte) ([]byte, error) {
	return args, nil
}

// test Register errors
func TestServSplitmodeErr(t *testing.T) {
	c := &Config{Sockname: "/tmp/test12.sock", Msglvl: "conn"}
	as, err := UnixServer(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// add a handler, successfully
	err = as.Register("echo", echo)
	if err != nil {
		t.Errorf("Couldn't add handler: %v", err)
	}
	// try to add 'echo' again
	err = as.Register("echo", echo)
	if err.Error() != "handler 'echo' already exists" {
		t.Errorf("Expected pre-existing handler 'echo' but got: %v", err)
	}
	as.Quit()
}

// implement an echo server
func TestServEchoNosplit(t *testing.T) {
	c := &Config{Sockname: "/tmp/test12.sock", Msglvl: "conn"}
	as, err := UnixServer(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.Register("echo", echo)
	as.Register("echonosplit", echonosplit)
	as.Register("echo nosplit", echonosplit)

	// launch echoclient. we should get a message about the
	// connection.
	go echosplitclient(as.s, t)
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if !strings.HasPrefix(msg.Txt, "client connected") {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
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

func echosplitclient(sn string, t *testing.T) {
	ac, err := pc.UnixClient(&pc.Config{Addr: sn})
	if err != nil {
		t.Fatalf("client instantiation failed! %s", err)
	}
	defer ac.Quit()

	// this one goes to a "argv" handler
	resp, err := ac.Dispatch([]byte("echo"), []byte("it works!"))
	if err != nil {
		t.Errorf("Error on read: '%v'", err)
	}
	if string(resp) != "it works!" {
		t.Errorf("Expected 'it works!' but got %v", resp)
	}
	// testing with JUST a command, no following args
	resp, err = ac.Dispatch([]byte("echo"), []byte(""))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "" {
		t.Errorf("Expected '' but got '%v'", string(resp))
	}
}
