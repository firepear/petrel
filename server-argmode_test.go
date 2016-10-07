package petrel

import (
	"strings"
	"testing"
)

// the echo function for our dispatch table
func echonosplit(args [][]byte) ([]byte, error) {
	return args[0], nil
}

// test Register errors
func TestServSplitmodeErr(t *testing.T) {
	c := &ServerConfig{Sockname: "/tmp/test12.sock", Msglvl: Conn}
	as, err := UnixServ(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// add a handler, successfully
	err = as.Register("echo", "args", echo)
	if err != nil {
		t.Errorf("Couldn't add handler: %v", err)
	}
	// now try to add a handler with an invalid mode
	err = as.Register("echonisplit", "nopesplit", echonosplit)
	if err.Error() != "invalid mode 'nopesplit'" {
		t.Errorf("Expected invalid mode 'nopesplit', but got: %v", err)
	}
	// finally, try to add 'echo' again
	err = as.Register("echo", "args", echo)
	if err.Error() != "handler 'echo' already exists" {
		t.Errorf("Expected pre-existing handler 'echo' but got: %v", err)
	}
	as.Quit()
}

// implement an echo server
func TestServEchoNosplit(t *testing.T) {
	c := &ServerConfig{Sockname: "/tmp/test12.sock", Msglvl: Conn}
	as, err := UnixServ(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.Register("echo", "args", echo)
	as.Register("echonosplit", "blob", echonosplit)
	as.Register("echo nosplit", "blob", echonosplit)

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
	ac, err := UnixClient(&ClientConfig{Addr: sn})
	if err != nil {
		t.Fatalf("client instantiation failed! %s", err)
	}
	defer ac.Close()

	// this one goes to a "args" handler
	resp, err := ac.Dispatch([]byte("echo it works!"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "it works!" {
		t.Errorf("Expected 'it works!' but got '%v'", string(resp))
	}
	// testing with JUST a command, no following args
	resp, err = ac.Dispatch([]byte("echo"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "" {
		t.Errorf("Expected '' but got '%v'", string(resp))
	}
	//and this one to a "blob" handler
	resp, err = ac.Dispatch([]byte("echonosplit it works!"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "it works!" {
		t.Errorf("Expected 'it works!' but got '%v'", string(resp))
	}
	// and this one to a handler with a quoted command (just to prove
	// out that functionality)
	resp, err = ac.Dispatch([]byte("'echo nosplit' it works!"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "it works!" {
		t.Errorf("Expected 'it works!' but got '%v'", string(resp))
	}
	// testing with JUST a command, no following args
	resp, err = ac.Dispatch([]byte("'echonosplit"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "" {
		t.Errorf("Expected '' but got '%v'", string(resp))
	}
}
