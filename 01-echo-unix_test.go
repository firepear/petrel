package petrel

import (
	"testing"

	"firepear.net/pclient"
)

// the echo function for our dispatch table
func echo(args [][]byte) ([]byte, error) {
	var bs []byte
	for i, arg := range args {
		bs = append(bs, arg...)
		if i != len(args) - 1 {
			bs = append(bs, byte(32))
		}
	}
	return bs, nil
}

// implement an echo server
func TestEchoServer(t *testing.T) {
	// instantiate petrel
	c := &Config{Sockname: "/tmp/test02.sock", Msglvl: All}
	as, err := NewUnix(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.AddHandler("echo", "split", echo)

	// launch a client and do some things
	go echoclient("/tmp/test02.sock", t)
	echotests(as, t)
	// shut down petrel
	as.Quit()
}

func echotests(as *Asock, t *testing.T) {
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
	if msg.Txt != "dispatching [echo]" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 101 {
		t.Errorf("msg.Code should have been 101 but got: %v", msg.Code)
	}
	// and a message that we have replied
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("successful cmd shouldn't be err, but got %v", msg.Err)
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
		t.Errorf("unsuccessful cmd shouldn't be err, but got %v", msg.Err)
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
}

// this time our (less) fake client will send a string over the
// connection and (hopefully) get it echoed back.
func echoclient(sn string, t *testing.T) {
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
	// for bonus points, let's send a bad command
	resp, err = ac.Dispatch([]byte("foo bar"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "Unknown command 'foo'. Available commands: echo " {
		t.Errorf("Expected bad command error but got '%v'", string(resp))
	}
}
