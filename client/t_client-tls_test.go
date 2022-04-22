package client

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"testing"

	ps "github.com/firepear/petrel/server"
)

var servertc *tls.Config
var clienttc *tls.Config

func init() {
	// set up client tls.Config (insecure because our test cert is
	// self-signed)
	certpem, err := os.ReadFile("../assets/cert.pem")
	key, err := os.ReadFile("../assets/privkey.pem")
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

func TestClientNewTLS(t *testing.T) {
	// instantiate unix petrel
	asconf := &ps.ServerConfig{Sockname: "127.0.0.1:10298", Msglvl: Fatal}
	as, err := ps.TLSServer(asconf, servertc)
	if err != nil {
		t.Errorf("Failed to create petrel instance: %v", err)
	}
	as.Register("echo", "blob", hollaback)
	// and now a client
	cconf := &ClientConfig{Addr: "127.0.0.1:10298"}
	c, err := TLSClient(cconf, clienttc)
	if err != nil {
		t.Errorf("Failed to create client: %v", err)
	}
	// and send a message
	resp, err := c.Dispatch([]byte("echo just the one test"))
	if err != nil {
		t.Errorf("Dispatch returned error: %v", err)
	}
	if string(resp) != "just the one test" {
		t.Errorf("Expected `just the one test` but got: `%v`", string(resp))
	}
	c.Quit()
	as.Quit()
}

func TestClientNewTLSFails(t *testing.T) {
	cconf := &ClientConfig{Addr: "999.255.255.255:10298"}
	c, err := TLSClient(cconf, clienttc)
	if err == nil {
		t.Errorf("Tried connecting to invalid IP but call succeeded: `%v`", c)
	}
}
