package asock

import (
	"net"
	"testing"
	"time"
)

// create an asocket with a one second timeout on its
// connections. connect to it with a client which does waits too long
// before trying to talk.
func TestConnTimeout(t *testing.T) {
	var d Dispatch
	// instantiate an asocket
	c := Config{"/tmp/test07.sock", 25, 32, All, nil}
	as, err := NewUnix(c, d)
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
	if msg.Txt != "ending session" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 197 {
		t.Errorf("msg.Code should be 197 but got: %v", msg.Code)
	}
	// shut down asocket
	as.Quit()
}

// the timeout on our connection is 1 second. we'll wait 1.2s then try
// to send/recv on it.
func sleeperclient(sn string, t *testing.T) {
	conn, err := net.Dial("unix", sn)
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	time.Sleep(50 * time.Millisecond)
	_, err = conn.Write([]byte("foo bar"))
	if err == nil {
		t.Error("conn should be closed due to timeout, but Write() succeeded")
	}
	res, err := readConn(conn)
	if err == nil {
		t.Errorf("Read should have failed due to timeout but got: %v", res)
	}
}

