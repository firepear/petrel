package asock

import (
	"net"
	"testing"
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
	conn, err := net.Dial("unix", sn)
	defer conn.Close()
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	conn.Write([]byte("echo it works!\n\n"))
	res, err := readConn(conn)
	if string(res) != "it works!\n\n" {
		t.Errorf("Expected 'it works!\\n\\n' but got '%v'", string(res))
	}
	// now try sending a second request
	_, err = conn.Write([]byte("foo bar\n\n"))
	if err == nil {
		t.Error("conn should be closed by one-shot server, but Write() succeeded")
	}
	res, err = readConn(conn)
	if err == nil {
		t.Errorf("Read should have failed byt got: %v", res)
	}
}
