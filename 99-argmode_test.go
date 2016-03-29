package asock

import (
	"testing"

	"firepear.net/pclient"
)

// the echo function for our dispatch table
func echonosplit(args [][]byte) ([]byte, error) {
	return args[0], nil
}

// test AddHandler errors
func TestSplitmodeErr(t *testing.T) {
	c := &Config{Sockname: "/tmp/test12.sock", Msglvl: Conn}
	as, err := NewUnix(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// add a handler, successfully
	err = as.AddHandler("echo", "split", echo)
	if err != nil {
		t.Errorf("Couldn't add handler: %v", err)
	}
	// now try to add a handler with an invalid argmode
	err = as.AddHandler("echonisplit", "nopesplit", echonosplit)
	if err.Error() != "invalid argmode 'nopesplit'" {
		t.Errorf("Expected invalid argmode 'nopesplit', but got: %v", err)
	}
	// finally, try to add 'echo' again
	err = as.AddHandler("echo", "split", echo)
	if err.Error() != "handler 'echo' already exists" {
		t.Errorf("Expected pre-existing handler 'echo' but got: %v", err)
	}
	as.Quit()
}

// implement an echo server
func TestEchoNosplit(t *testing.T) {
	c := &Config{Sockname: "/tmp/test12.sock", Msglvl: Conn}
	as, err := NewUnix(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.AddHandler("echo", "split", echo)
	as.AddHandler("echonosplit", "nosplit", echonosplit)
	as.AddHandler("echo nosplit", "nosplit", echonosplit)

	// launch echoclient. we should get a message about the
	// connection.
	go echosplitclient(as.s, t)
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "client connected" {
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
	// shut down asocket
	as.Quit()
}

func echosplitclient(sn string, t *testing.T) {
	ac, err := pclient.NewUnix(&pclient.Config{Addr: sn})
	if err != nil {
		t.Fatalf("pclient instantiation failed! %s", err)
	}
	defer ac.Close()

	// this one goes to a "split" handler
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
	//and this one to a "nosplit" handler
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
