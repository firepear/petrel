package petrel

import (
	"strings"
	"testing"
)

// implement an echo server
func TestServPlenex(t *testing.T) {
	// instantiate petrel
	c := &ServerConfig{Sockname: "/tmp/test05c.sock", Msglvl: All, Reqlen: 10}
	as, err := UnixServer(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.Register("echo", "argv", echo)

	// launch a client and do some things
	go reqclient("/tmp/test05c.sock", t)
	reqtests(as, t)
	// shut down petrel
	as.Quit()
}

func reqtests(as *Server, t *testing.T) {
	// we should get a message about the connection.
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if !strings.HasPrefix(msg.Txt, "client connected") {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// and a message about dispatching the command
	msg = <-as.Msgr
	if msg.Txt != perrs["plenex"].Txt {
		t.Errorf("expected %s but got: %s", perrs["plenex"].Txt, msg.Txt)
	}
	if msg.Code != 402 {
		t.Errorf("msg.Code should have been 402 but got: %v", msg.Code)
	}
}

// this time our (less) fake client will send a string over the
// connection and (hopefully) get it echoed back.
func reqclient(sn string, t *testing.T) {
	ac, err := UnixClient(&ClientConfig{Addr: sn})
	if err != nil {
		t.Fatalf("client instantiation failed! %s", err)
	}
	defer ac.Quit()

	resp, err := ac.Dispatch([]byte("echo this string is way too long! it won't work!"))
	if len(resp) != 1 && resp[0] != 255 {
		t.Errorf("len resp should 1 & resp[0] should be 255, but got len %d and '%v'", len(resp), string(resp))
	}
	if err.(*Perr).Code != perrs["plenex"].Code {
		t.Errorf("err.Code should be %d but is %v", perrs["plenex"].Code, err.(*Perr).Code)
	}
	if err.(*Perr).Txt != perrs["plenex"].Txt {
		t.Errorf("err.Txt should be %s but is %v", perrs["plenex"].Txt, err.(*Perr).Txt)
	}
}
