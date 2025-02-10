package server

import (
	"errors"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	p "github.com/firepear/petrel"
)

// create and destroy an idle petrel instance
func TestServStartStop(t *testing.T) {
	// fail to instantiate petrel by using a terrible filename
	c := &Config{Sockname: "zzz/zzz/zzz/zzz", Msglvl: "debug"}
	as, err := UnixServer(c, 700)
	if err == nil {
		t.Error("that should have failed, but didn't")
	}

	// instantiate petrel
	c = &Config{Sockname: "/tmp/test00.sock", Msglvl: "debug"}
	as, err = UnixServer(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// stat it
	fi, err := os.Stat(as.s)
	if err != nil {
		t.Errorf("Couldn't stat socket: %v", err)
	}
	fm := fi.Mode()
	if fm&os.ModeSocket == 1 {
		t.Errorf("'Socket' is not a socket %v", fm)
	}
	as.Quit()
}

// create petrel. connect to it with a client which does
// nothing but wait 1/10 second before disconnecting. tear down
// petrel.
func TestServConnServer(t *testing.T) {
	// instantiate petrel
	c := &Config{Sockname: "/tmp/test01.sock", Msglvl: "debug"}
	as, err := UnixServer(c, 700)
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
	if !strings.HasPrefix(msg.Txt, "client connected") {
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
	// shut down petrel
	as.Quit()
}

// these tests check for petrel.Msg implementing the Error interface
// properly.
func TestServMsgError(t *testing.T) {
	c := &Config{Sockname: "/tmp/test13.sock", Msglvl: "debug"}
	as, err := UnixServer(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}

	// first Msg: bare bones
	as.genMsg(1, 1, p.Stats["success"], "", nil)
	m := <-as.Msgr
	s := m.Error()
	if s != "conn 1 req 1 status 200 (reply sent)" {
		t.Errorf("Expected 'conn 1 req 1 status 200 (reply sent)' but got '%v'", s)
	}

	// now with Msg.Txt
	as.genMsg(1, 1, p.Stats["success"], "foo", nil)
	m = <-as.Msgr
	s = m.Error()
	if s != "conn 1 req 1 status 200 (reply sent: [foo])" {
		t.Errorf("Expected 'conn 1 req 1 status 200 (reply sent: [foo])' but got '%v'", s)
	}

	// and an error
	e := errors.New("something bad")
	as.genMsg(1, 1, p.Stats["success"], "foo", e)
	m = <-as.Msgr
	s = m.Error()
	if s != "conn 1 req 1 status 200 (reply sent: [foo]); err: something bad" {
		t.Errorf("Expected 'conn 1 req 1 status 200 (reply sent: [foo]); err: something bad' but got '%v'", s)
	}
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
