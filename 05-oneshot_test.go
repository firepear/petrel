package adminsock

import (
	"net"
	"testing"
)

// function readConn() is defined in test 02.

func TestOneShot(t *testing.T) {
	d := make(Dispatch) // create Dispatch
	d["echo"] = echo    // and put a function in it
	//instantiate an adminsocket which will spawn connections that
	//close after one response
	as, err := New("test05", d, -1, All)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
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
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection drop should be nil, but got %v", err)
	}
	if msg.Txt != "disconnected one-shot session" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 198 {
		t.Errorf("msg.Code should be 198 but is %v", msg.Code)
	}		
	// shut down adminsocket
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
	conn.Write([]byte("echo it works!"))
	res, err := readConn(conn)
	if string(res) != "it works!" {
		t.Errorf("Expected 'it works!' but got '%v'", string(res))
	}
	// now try sending a second request
	_, err = conn.Write([]byte("foo bar"))
	if err == nil {
		t.Error("conn should be closed by one-shot server, but Write() succeeded")
	}
	res, err = readConn(conn)
	if err == nil {
		t.Errorf("Read should have failed byt got: %v", res)
	}
}

