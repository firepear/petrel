package adminsock

import (
	"net"
	"testing"
	"time"
)

// create an adminsocket with a one second timeout on its
// connections. connect to it with a client which does waits too long
// before trying to talk.
func TestConnTimeout(t *testing.T) {
	var d Dispatch
	// instantiate an adminsocket
	as, err := New("test07", d, 1, All)
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
	if msg.Txt != "adminsock conn 1 opened" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr
	if msg.Err == nil {
		t.Errorf("connection drop should be an err, but got nil")
	}
	if msg.Txt != "adminsock conn 1 client lost" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// shut down adminsocket
	as.Quit()
}

// the timeout on our connection is 1 second. we'll wait 1.2s then try
// to send/recv on it.
func sleeperclient(sn string, t *testing.T) {
	conn, err := net.Dial("unix", sn)
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	time.Sleep(1200 * time.Millisecond)
	_, err = conn.Write([]byte("foo bar"))
	if err == nil {
		t.Error("conn should be closed due to timeout, but Write() succeeded")
	}
	res, err := readConn(conn)
	if err == nil {
		t.Errorf("Read should have failed due to timeout but got: %v", res)
	}
}

