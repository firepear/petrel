package petrel

import (
	"fmt"
	"strings"
	"testing"
)

// the faulty echo function for our dispatch table
func badecho(s [][]byte) ([]byte, error) {
	return nil, fmt.Errorf("oh no something is wrong")
}

// implement an echo server with a bad command
func TestServInternalError(t *testing.T) {
	// instantiate petrel
	c := &ServerConfig{Sockname: "/tmp/test08.sock", Msglvl: All}
	as, err := UnixServer(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.Register("echo", "argv", echo)
	as.Register("badecho", "argv", badecho)

	// launch echoclient
	go internalerrclient(as.s, t)
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if !strings.HasPrefix(msg.Txt, "client connected") {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	msg = <-as.Msgr // ignore msg on dispatch
	msg = <-as.Msgr // ignore msg of reply sent
	// wait for msg from unsuccessful command
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("unsuccessful cmd shouldn't be err, but got %v", err)
	}
	if msg.Txt != "dispatching: [badecho]" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	msg = <-as.Msgr
	if msg.Txt != "request failed" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	if msg.Err.Error() != "oh no something is wrong" {
		t.Errorf("unsuccessful cmd should be an error, but got %v", msg.Err)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr
	if msg.Err == nil {
		t.Errorf("connection drop should be an err, but got nil")
	}
	if msg.Txt != "client disconnected" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// shut down petrel
	as.Quit()
}

// this time our (less) fake client will send a string over the
// connection and (hopefully) get it echoed back.
func internalerrclient(sn string, t *testing.T) {
	ac, err := UnixClient(&ClientConfig{Addr: sn})
	if err != nil {
		t.Fatalf("client instantiation failed! %s", err)
	}
	defer ac.Quit()

	resp, err := ac.Dispatch([]byte("echo it works!"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "it works!" {
		t.Errorf("Expected 'it works!' but got '%v'", string(resp))
	}
	// it's not going to work this time though :(
	resp, err = ac.Dispatch([]byte("badecho foo bar"))
	if len(resp) != 1 && resp[0] != 255 {
		t.Errorf("len resp should 1 & resp[0] should be 255, but got len %d and '%v'", len(resp), string(resp))
	}
	if err.(*Perr).Code != perrs["reqerr"].Code {
		t.Errorf("err.Code should be %d but is %v", perrs["reqerr"].Code, err.(*Perr).Code)
	}
	if err.(*Perr).Txt != perrs["reqerr"].Txt {
		t.Errorf("err.Txt should be %s but is %v", perrs["reqerr"].Txt, err.(*Perr).Txt)
	}
}
