package petrel

import (
	"crypto/tls"
	"crypto/x509"
	"strings"
	"testing"
)

// this file tests Asock with a TLS connection. The following keys are
// good for 10 years from mid-May 2015.

var servertc *tls.Config
var clienttc *tls.Config

const keyPEM = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA8bfvIhxFRD5kuJV3KzN1WOwRjYwIPFM2xSnHsBHG9gY4+P54
rnalgME0xPpThHP4XMbJ1Njn2IlJak032cIFIjgfODEgNH6W4ndBNdpRqLRIyOfz
zNRx2c2KVf6SwUw9CghxQ5fgL2GHU5XO5EYCeB0oXiHo4DDp564NPCfLElLB4bBv
rCel54mZk/8IHzjCN2r3vEU2Ad9o1hJaOEbd9weRIzmozuOG4XRPjwkLmvS8c6Jm
xg0Bb0EfI/h5kiZlkj0KUtBd8SKqHkshM3v+l2ddzQF8jNXsnWin8xSNuezlHuBv
LsPc7+GLUp1FKlZgA9WlVo6hetbO7hGE+CElvwIDAQABAoIBAGk4F/BRPhWm01FG
PsmfbMV4fWuQOUWJM54/wZzzIBiYPNSmcQIAw6p4b/AOx6wwjzxTjCgLA2FO4ZZU
ZqtzuahbpbtgJxSyxhturgQzNLirQcOytH3FPIoC3uTwHBHojemAI025Hu2BFtdb
ruPPVePTTW8sc6KjqC4hpcE50Tv30bUr7Vbtz0qC2cYUaGyCt734PY21pk9qTQRY
n7DSlT+eare9EqSZAAjqPjAZrQeW+rjB/d0sY7oOB/4iFa8tEN/SMCe9GrjHInG6
g0ltV0EMkWqJtkb7sGm0wFxJugGtRmZu6lKvVXl1UZh7rVJ+f2EUrGypDw5aCeNw
3Bq+1oECgYEA9NQkIo5p0btvl3kJoDe/Ggo5NNnotwMzpFqldcPQjakamVlNaYZ3
eBieZP/f5UH8/4Qe0HwE15FoIATjvksX1D9x2TYx6tATYN/rIV0uAM8vc+PoMQyI
r+bxACLwu6jt8ZW+d5/+xnbC1l08xriprV6kz5EqoTOq5NMkhGJPvasCgYEA/L92
dG5VGrzz6+oDxvHbQIHDGkUdSRuGU1rIUMWQPzLmagWU28SoCCmooBfG0w+QModO
QpcvWjHTPC00bZcmHTJRS0wq/aEdZk989QiNKN651HQgUqyBk6w6HRttkDSJuA+e
J8kem33C3HZVrGfOIn8bLVeBCtYYBtfsAFZm3D0CgYBllP/JNr3BP7v1ZUsRJxAr
hcJmo2NjS/jJYLL2QeDuZhObPOpZtmkrc0uFLIWBYffPLMp8Rnjb2IETh/PWqOGi
NxDNxya+/saLk1zD4x2LSGuv8ggNEd5E3dVw8Q5hTp4rdq4ohEH5pp0AxH7LFSOR
w4sudFTzvbRSbSjhpMjhMQKBgQCQoAprM1s1cpvtCbphk7GPJvF6TcQlOj/R4Kex
OGuDDmA0mL8GRnCUQyo/eXuG+Gfd0fjhN9ubs9kYnRFcCFqB6HIGMS6EdTX6fk+V
cvA7S79wJ4b7Z8S5uJqEX1aBZt7LWPx57aa6+OqQ9pGtlrSonqzxdBneFoYnHFTq
GIbBTQKBgENNq0u1i0TI7mlE+C0HSL4S6k00gdJm5oXKcJ7GZDO8xUXXsulvrtkT
qMmtqiVIhMrBy+gSnjvON/06THx5Anz4X2cEQM5rF0YYrVS4YhIbX4UUQLYKvNJB
so8iKX3DM3flyJSyGd9HgLJDRLWR0Ic373O1+++cjgOG1aKBYFHs
-----END RSA PRIVATE KEY-----`
const certPEM = `
-----BEGIN CERTIFICATE-----
MIIC+jCCAeSgAwIBAgIQRKnue+k484dj8UIiIK6bsTALBgkqhkiG9w0BAQswEjEQ
MA4GA1UEChMHQWNtZSBDbzAeFw0xNTAxMDEwMDAwMDFaFw0yNDAzMTkwMDAwMDFa
MBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQDxt+8iHEVEPmS4lXcrM3VY7BGNjAg8UzbFKcewEcb2Bjj4/niudqWAwTTE
+lOEc/hcxsnU2OfYiUlqTTfZwgUiOB84MSA0fpbid0E12lGotEjI5/PM1HHZzYpV
/pLBTD0KCHFDl+AvYYdTlc7kRgJ4HSheIejgMOnnrg08J8sSUsHhsG+sJ6XniZmT
/wgfOMI3ave8RTYB32jWElo4Rt33B5EjOajO44bhdE+PCQua9LxzombGDQFvQR8j
+HmSJmWSPQpS0F3xIqoeSyEze/6XZ13NAXyM1eydaKfzFI257OUe4G8uw9zv4YtS
nUUqVmAD1aVWjqF61s7uEYT4ISW/AgMBAAGjUDBOMA4GA1UdDwEB/wQEAwIApDAT
BgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MBYGA1UdEQQPMA2C
BVs6OjFdhwR/AAABMAsGCSqGSIb3DQEBCwOCAQEAEvRIkTomdLZw8RJJHjGtmmyB
2NS/S6tYcHJKK+nAZ+AsLxB4BXR9+obLP1vUqFpLXswxrKIv7pb7ZsmWWn1enFJM
jSLAH6mIFSeoK538rKGCXAHBly5yIhNTQlKdFPkqo3km8Nw89FvDY5xjf0vqlADZ
V++hoMoOVRQTmE1OUiWzLgNhFYHfTo5q1DiwoD/JaQDgzJQoDeo8m35HiqKplc1h
4g9Q3yjjeloXu/mOtcXIpnElKc5m4vyyyBloe9xDWwDIzpYRd4AJJPzeXxZOr7zo
C9JwGMXEovVjdaJeBhkm9sv2lsOO1MBIxYzYUUs0F5jj5aiod72XHTGm7j1Vgg==
-----END CERTIFICATE-----`

func init() {
	// set up client tls.Config (insecure because our test cert is
	// self-signed)
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(certPEM))
	if !ok {
		panic("failed to parse root certificate")
	}
	clienttc = &tls.Config{RootCAs: roots, InsecureSkipVerify: true}
	// set up server tls.Config
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
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
	c := &ServerConfig{Sockname: "127.0.0.1:50707", Msglvl: All}
	as, err := TLSServ(c, clienttc)
	if err == nil {
		as.Quit()
		t.Errorf("tls.Listen with client config shouldn't have worked, but did")
	}

	// instantiate petrel (success)
	c = &ServerConfig{Sockname: "127.0.0.1:50707", Msglvl: All}
	as, err = TLSServ(c, servertc)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	if as.s != "127.0.0.1:50707" {
		t.Errorf("Socket name should be '127.0.0.1:50707' but got '%v'", as.s)
	}
	as.Register("echo", "argv", echo)

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
	c := &ServerConfig{Sockname: "[::1]:50707", Msglvl: All}
	as, err := TLSServ(c, servertc)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	if as.s != "[::1]:50707" {
		t.Errorf("Socket name should be '[::1]:50707' but got '%v'", as.s)
	}
	as.Register("echo", "argv", echo)

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
	ac, err := TLSClient(&ClientConfig{Addr: sn}, clienttc)
	if err != nil {
		t.Fatalf("client instantiation failed! %s", err)
	}
	defer ac.Quit()

	resp, err := ac.Dispatch([]byte("echo it works!"))
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(resp) != "it works!" {
		t.Errorf("Expected 'it works!\\n\\n' but got '%v'", string(resp))
	}
}

