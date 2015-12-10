package asock

import (
	"testing"
	"time"

	"firepear.net/aclient"
)

// create an asocket with a one second timeout on its
// connections. connect to it with a client which does waits too long
// before trying to talk.
func TestConnTimeout(t *testing.T) {
	// instantiate an asocket
	c := Config{Sockname: "/tmp/test07.sock", Timeout: 25, Msglvl: All}
	as, err := NewUnix(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}

	// launch fakeclient. we should get a message about the
	// connection.
	go sleeperclient(as.s, t)
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
	if msg.Txt != "failed to read mlen from socket" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 196 {
		t.Errorf("msg.Code should be 197 but got: %v", msg.Code)
	}
	// shut down asocket
	as.Quit()
}

// the timeout on our connection is 25ms. we'll wait 50ms then try
// to send/recv on it.
func sleeperclient(sn string, t *testing.T) {
	ac, err := aclient.NewUnix(aclient.Config{Addr: sn})
	if err != nil {
		t.Fatalf("aclient instantiation failed! %s", err)
	}
	defer ac.Close()

	time.Sleep(50 * time.Millisecond)
	resp, err := ac.Dispatch([]byte("echo it works!"))
	if err == nil {
		t.Error("conn should be closed due to timeout, but Write() succeeded")
	}
	if resp != nil {
		t.Errorf("Read should have failed due to timeout but got: %v", resp)
	}
}

