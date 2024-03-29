package server

import (
	"strings"
	"testing"

	p "github.com/firepear/petrel"
	pc "github.com/firepear/petrel/client"
)

// the echo function for our dispatch table, and readConn for the
// client, are defined in test02

// implement an echo server
func TestServHMACTCPServerer(t *testing.T) {
	// instantiate petrel
	c := &Config{Sockname: "127.0.0.1:50711", Msglvl: "debug", HMACKey: []byte("test")}
	as, err := TCPServer(c)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// load the echo func into the dispatch table
	err = as.Register("echo", echo)
	if err != nil {
		t.Errorf("Couldn't add handler func: %v", err)
	}

	// launch echoclient. we should get a message about the
	// connection.
	go echoHMACTCPclient(as.s, t)
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if !strings.HasPrefix(msg.Txt, "client connected") {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// and a message about dispatching the command
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("successful cmd shouldn't be err, but got %v", err)
	}
	if msg.Txt != "dispatching: [echo]" {
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

	// now it's going to happen again, but there's going to be an
	// HMAC mismatch.
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if !strings.HasPrefix(msg.Txt, "client connected") {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// hmac mismatch will cause an immediate connection close
	msg = <-as.Msgr
	if msg.Code != p.Errs["badmac"].Code {
		t.Errorf("msg.Code should have been %d but got: %d", p.Errs["badmac"].Code, msg.Code)
	}
	// shut down petrel
	as.Quit()
}

func echoHMACTCPclient(sn string, t *testing.T) {
	// Matching HMAC keys should work
	ac, err := pc.TCPClient(&pc.Config{Addr: sn, HMACKey: []byte("test")})
	if err != nil {
		t.Fatalf("client instantiation failed! %s", err)
	}
	defer ac.Quit()

	resp, err := ac.Dispatch([]byte("echo"), []byte("it works!"))
	if err != nil {
		t.Errorf("Dispatch error: %v", err)
	}
	if string(resp) != "it works!" {
		t.Errorf("Expected 'it works!' but got '%v'", string(resp))
	}

	// HMAC mismatch should fail
	ac, err = pc.TCPClient(&pc.Config{Addr: sn, HMACKey: []byte("terp")})
	if err != nil {
		t.Fatalf("client instantiation failed! %s", err)
	}
	defer ac.Quit()

	_, err = ac.Dispatch([]byte("echo"), []byte("it works!"))
	if err == nil {
		t.Errorf("HMAC mismatch should have sent back an error, but got nil")
	}
	if err.(*p.Perr).Code != p.Errs["badmac"].Code {
		t.Errorf("err.Code should be %d but is %v", p.Errs["badmac"].Code, err.(*p.Perr).Code)
	}
	if err.(*p.Perr).Txt != p.Errs["badmac"].Txt {
		t.Errorf("err.Txt should be %s but is %v", p.Errs["badmac"].Txt, err.(*p.Perr).Txt)
	}
}
