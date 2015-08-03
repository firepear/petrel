package asock

import (
	"net"
	"testing"
)

// the echo function for our dispatch table, and readConn for the
// client, are defined in test02

// implement an echo server
func TestEchoTCPServer(t *testing.T) {
	// instantiate an asocket (failure)
	c := Config{Sockname: "127.0.0.1:1", Msglvl: All}
	as, err := NewTCP(c)
	if err == nil {
		as.Quit()
		t.Errorf("Tried to listen on an impossible IP, but it worked")
	}

	// instantiate an asocket
	c = Config{Sockname: "127.0.0.1:50709", Msglvl: All}
	as, err = NewTCP(c)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	if as.s != "127.0.0.1:50709" {
		t.Errorf("Socket name should be '127.0.0.1:50709' but got '%v'", as.s)
	}
	// load the echo func into the dispatch table
	err = as.AddHandler("echo", "nosplit", echo)
	if err != nil {
		t.Errorf("Couldn't add handler func: %v", err)
	}
	if len(as.d) != 1 {
		t.Errorf("as.d should be len 1, but got %v", len(as.d))
	}
	if _, ok := as.d["echo"]; !ok {
		t.Errorf("Can't find dispatch function 'echo'")
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
	// wait for msg from nil command
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("nil cmd shouldn't be err, but got %v", err)
	}
	if msg.Txt != "nil request" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 401 {
		t.Errorf("msg.Code should have been 401 but got: %v", msg.Code)
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
	// instantiate an asocket
	c := Config{Sockname: "[::1]:50709", Msglvl: All}
	as, err := NewTCP(c)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	if as.s != "[::1]:50709" {
		t.Errorf("Socket name should be '[::1]:50709' but got '%v'", as.s)
	}
	// load the echo func into the dispatch table, with argmode of
	// split this time
	err = as.AddHandler("echo", "split", echo)
	if err != nil {
		t.Errorf("Couldn't add handler func: %v", err)
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
	// wait for msg from nil command
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("nil cmd shouldn't be err, but got %v", err)
	}
	if msg.Txt != "nil request" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Code != 401 {
		t.Errorf("msg.Code should have been 401 but got: %v", msg.Code)
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
	conn.Write([]byte("echo it works!\n\n"))
	res, err := readConn(conn)
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "it works!\n\n" {
		t.Errorf("Expected 'it works!\\n\\n' but got '%v'", string(res))
	}
	// for bonus points, let's send a bad command
	conn.Write([]byte("foo bar\n\n"))
	res, err = readConn(conn)
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "Unknown command 'foo'. Available commands: echo \n\n" {
		t.Errorf("Expected bad command error but got '%v'", string(res))
	}
	// and a null command!
	conn.Write([]byte("\n\n"))
	res, err = readConn(conn)
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "Received empty request. Available commands: echo \n\n" {
		t.Errorf("Expected bad command error but got '%v'", string(res))
	}
}

