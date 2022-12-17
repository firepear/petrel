package server

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"
	"testing"

	pc "github.com/firepear/petrel/client"
)

// this file tests Asock with a TLS connection. The following keys are
// good for 10 years from mid-May 2015.

var servertc *tls.Config
var clienttc *tls.Config

func init() {
	// set up client tls.Config (insecure because our test cert is
	// self-signed)
	certpem, _ := os.ReadFile("../assets/cert.pem")
	key, _ := os.ReadFile("../assets/privkey.pem")
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(certpem)
	if !ok {
		panic("failed to parse root certificate")
	}
	clienttc = &tls.Config{RootCAs: roots, InsecureSkipVerify: true}
	// set up server tls.Config
	cert, err := tls.X509KeyPair(certpem, key)
	if err != nil {
		panic("failed to generate x509 keypair")
	}
	servertc = &tls.Config{Certificates: []tls.Certificate{cert}}
}

// the echo function for our dispatch table, and readConn for the
// client, are defined in test02

// implement an echo server
func TestServEchoTLSServer(t *testing.T) {
	// instantiate petrel (failure)
	c := &Config{Sockname: "127.0.0.1:50707", Msglvl: "debug"}
	as, err := TLSServer(c, clienttc)
	if err == nil {
		as.Quit()
		t.Errorf("tls.Listen with client config shouldn't have worked, but did")
	}

	// instantiate petrel (success)
	c = &Config{Sockname: "127.0.0.1:50707", Msglvl: "debug"}
	as, err = TLSServer(c, servertc)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	if as.s != "127.0.0.1:50707" {
		t.Errorf("Socket name should be '127.0.0.1:50707' but got '%v'", as.s)
	}
	as.Register("echo", echo)

	// launch echoclient. we should get a message about the
	// connection.
	go echoTLSclient(as.s, t)
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

// now do it in ipv6
func TestServEchoTLS6Server(t *testing.T) {
	// instantiate petrel
	c := &Config{Sockname: "[::1]:50707", Msglvl: "debug"}
	as, err := TLSServer(c, servertc)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	if as.s != "[::1]:50707" {
		t.Errorf("Socket name should be '[::1]:50707' but got '%v'", as.s)
	}
	as.Register("echo", echo)

	// launch echoclient. we should get a message about the
	// connection.
	go echoTLSclient(as.s, t)
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
func echoTLSclient(sn string, t *testing.T) {
	ac, err := pc.TLSClient(&pc.Config{Addr: sn}, clienttc)
	if err != nil {
		t.Fatalf("client instantiation failed! %s", err)
	}
	defer ac.Quit()

	resp, err := ac.Dispatch([]byte("echo"), []byte("it works!"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "it works!" {
		t.Errorf("Expected 'it works!\\n\\n' but got '%v'", string(resp))
	}
}
