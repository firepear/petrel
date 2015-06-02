package client

import (
	"testing"
	"firepear.net/asock"
)

func hollaback(args [][]byte) ([]byte, error) {
	return args[0], nil
}

func TestNewUnix(t *testing.T) {
	// instantiate unix asock
	asdisp := make(asock.Dispatch)
	asdisp["echo"] = &asock.DispatchFunc{hollaback, "nosplit"}
	asconf := asock.Config{Sockname: "/tmp/clienttest.sock", Msglvl: asock.Fatal}
	as, err := asock.NewUnix(asconf, asdisp, 700)
	if err != nil {
		t.Errorf("Failed to create asock instance: %v", err)
	}
	// and now a client
	cconf := Config{Addr: "/tmp/clienttest.sock"}
	c, err := NewUnix(cconf)
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

func TestNewTCP(t *testing.T) {
	// instantiate unix asock
	asdisp := make(asock.Dispatch)
	asdisp["echo"] = &asock.DispatchFunc{hollaback, "nosplit"}
	asconf := asock.Config{Sockname: "127.0.0.1:10298", Msglvl: asock.Fatal}
	as, err := asock.NewTCP(asconf, asdisp)
	if err != nil {
		t.Errorf("Failed to create asock instance: %v", err)
	}
	// and now a client
	cconf := Config{Addr: "127.0.0.1:10298"}
	c, err := NewTCP(cconf)
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

func TestNewTCPFails(t *testing.T) {
	cconf := Config{Addr: "999.255.255.255:10298"}
	c, err := NewTCP(cconf)
	if err == nil {
		t.Errorf("Tried connecting to invalid IP but call succeeded: `%v`", c)
	}
}

func TestNewUnixFails(t *testing.T) {
	cconf := Config{Addr: "/foo/999.255.255.255"}
	c, err := NewUnix(cconf)
	if err == nil {
		t.Errorf("Tried connecting to invalid path but call succeeded: `%v`", c)
	}
}
