package asock

import (
	"net"
	"testing"
)

// the echo function for our dispatch table, and readConn for the
// client, are defined in test02

// implement an echo server
func TestEchoTCPServer(t *testing.T) {
	d := make(Dispatch) // create Dispatch
	d["echo"] = &DispatchFunc{echo, "split"} // and put a function in it

	// instantiate an asocket (failure)
	c := Config{"127.0.0.1:1", 0, 0, All, nil}
	as, err := NewTCP(c, d)
	if err == nil {
		as.Quit()
		t.Errorf("Tried to listen on an impossible IP, but it worked")
	}

	// instantiate an asocket
	c = Config{"127.0.0.1:50707", 0, 0, All, nil}
	as, err = NewTCP(c, d)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	if as.s != "127.0.0.1:50707" {
		t.Errorf("Socket name should be '127.0.0.1:50707' but got '%v'", as.s)
	}
	// launch echoclient. we should get a message about the
	// connection.
	go echoTCPclient(as.s, t)
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
		t.Errorf("successful cmd shouldn't be err, but got %v", err)
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
		t.Errorf("successful cmd shouldn't be err, but got %v", err)
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
		t.Errorf("unsuccessful cmd shouldn't be err, but got %v", err)
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
	// shut down asocket
	as.Quit()
}

// now do it in ipv6
func TestEchoTCP6Server(t *testing.T) {
	d := make(Dispatch) // create Dispatch
	d["echo"] = &DispatchFunc{echo, "split"} // and put a function in it
	// instantiate an asocket
	c := Config{"[::1]:50707", 0, 0, All, nil}
	as, err := NewTCP(c, d)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	if as.s != "[::1]:50707" {
		t.Errorf("Socket name should be '[::1]:50707' but got '%v'", as.s)
	}
	// launch echoclient. we should get a message about the
	// connection.
	go echoTCPclient(as.s, t)
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
		t.Errorf("successful cmd shouldn't be err, but got %v", err)
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
		t.Errorf("successful cmd shouldn't be err, but got %v", err)
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
		t.Errorf("unsuccessful cmd shouldn't be err, but got %v", err)
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
	// shut down asocket
	as.Quit()
}

// this time our (less) fake client will send a string over the
// connection and (hopefully) get it echoed back.
func echoTCPclient(sn string, t *testing.T) {
	conn, err := net.Dial("tcp", sn)
	defer conn.Close()
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	conn.Write([]byte("echo it works!"))
	res, err := readConn(conn)
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "it works!" {
		t.Errorf("Expected 'it works!' but got '%v'", string(res))
	}
	// for bonus points, let's send a bad command
	conn.Write([]byte("foo bar"))
	res, err = readConn(conn)	
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "Unknown command 'foo'\nAvailable commands:\n    echo\n" {
		t.Errorf("Expected 'it works!' but got '%v'", string(res))
	}
}
