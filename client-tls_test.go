package petrel

import (
	"crypto/x509"
	"crypto/tls"
	"testing"
)

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

func TestNewTLS(t *testing.T) {
	// instantiate unix petrel
	asconf := &ServerConfig{Sockname: "127.0.0.1:10298", Msglvl: Fatal}
	as, err := TLSServ(asconf, servertc)
	if err != nil {
		t.Errorf("Failed to create petrel instance: %v", err)
	}
	as.AddFunc("echo", "blob", hollaback)
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
	c.Close()
	as.Quit()
}

func TestNewTLSFails(t *testing.T) {
	cconf := &ClientConfig{Addr: "999.255.255.255:10298"}
	c, err := TLSClient(cconf, clienttc)
	if err == nil {
		t.Errorf("Tried connecting to invalid IP but call succeeded: `%v`", c)
	}
}
