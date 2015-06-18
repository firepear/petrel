package asock

import (
	"net"
	"testing"
	"time"
)

// create an asocket. connect to it with a client which does
// nothing but wait 1/10 second before disconnecting. tear down
// asocket.
func TestConnHandler(t *testing.T) {
	// instantiate an asocket
	c := Config{
		Sockname: "/tmp/test01.sock",
		Msglvl: All,
	}
	as, err := NewUnix(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch fakeclient. we should get a message about the
	// connection.
	go fakeclient(as.s, t)
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "client connected" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Conn != 1 {
		t.Errorf("msg.Conn should be 1 but got: %v", msg.Conn)
	}
	if msg.Req != 0 {
		t.Errorf("msg.Req should be 0 but got: %v", msg.Req)
	}
	if msg.Code != 100 {
		t.Errorf("msg.Code should be 100 but got: %v", msg.Code)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr
	if msg.Err == nil {
		t.Errorf("connection drop should be an err, but got nil")
	}
	if msg.Txt != "client disconnected" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 198 {
		t.Errorf("msg.Code should be 198 but got: %v", msg.Code)
	}
	// shut down asocket
	as.Quit()
}

// we need a fake client in order to test here. but it can be really,
// really fake. we're not even going to test send/recv yet.
func fakeclient(sn string, t *testing.T) {
	conn, err := net.Dial("unix", sn)
	defer conn.Close()
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	time.Sleep(100 * time.Millisecond)
}
