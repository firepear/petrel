package asock

import (
	"testing"

	"firepear.net/aclient"
)

// function readConn() is defined in test 02.

func TestOneShot(t *testing.T) {
	//instantiate an asocket which will spawn connections that
	//close after one response
	c := Config{Sockname: "/tmp/test05.sock", Timeout: -100, Msglvl: All}
	as, err := NewUnix(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.AddHandler("echo", "split", echo)

	// launch oneshotclient.
	go oneshotclient(as.s, t)
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "client connected" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// wait for disconnect Msg
	<-as.Msgr // discard cmd dispatch message
	<-as.Msgr // discard reply sent message
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection drop should be nil, but got %v", err)
	}
	if msg.Txt != "ending session" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 197 {
		t.Errorf("msg.Code should be 197 but is %v", msg.Code)
	}
	// shut down asocket
	as.Quit()
}

// this time our (less) fake client will send a string over the
// connection and (hopefully) get it echoed back.
func oneshotclient(sn string, t *testing.T) {
	ac, err := aclient.NewUnix(aclient.Config{Addr: sn})
	if err != nil {
		t.Fatalf("aclient instantiation failed! %s", err)
	}
	defer ac.Close()

	resp, err := ac.Dispatch([]byte("echo it works!"))
	if string(resp) != "it works!" {
		t.Errorf("Expected 'it works!\\n\\n' but got '%v'", string(resp))
	}
	// now try sending a second request
	resp, err = ac.Dispatch([]byte("echo it works!"))
	if err == nil {
		t.Error("conn should be closed by one-shot server, but Write() succeeded")
	}
	if resp != nil {
		t.Errorf("Read should have failed byt got: %v", resp)
	}
}
