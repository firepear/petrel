package client

import (
	"testing"
	"firepear.net/petrel/server"
)

func hollaback(args [][]byte) ([]byte, error) {
	return args[0], nil
}

func TestNewUnix(t *testing.T) {
	// instantiate unix petrel
	asconf := &server.Config{Sockname: "/tmp/clienttest.sock", Msglvl: petrel.Fatal}
	as, err := server.NewUnix(asconf, 700)
	if err != nil {
		t.Errorf("Failed to create petrel instance: %v", err)
	}
	err = as.AddFunc("echo", "blob", hollaback)
	if err != nil {
		t.Errorf("Failed to add func: %v", err)
	}
	// and now a client
	cconf := &Config{Addr: "/tmp/clienttest.sock"}
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
	// instantiate unix petrel
	asconf := &server.Config{Sockname: "127.0.0.1:10298", Msglvl: petrel.Fatal}
	as, err := server.NewTCP(asconf)
	if err != nil {
		t.Errorf("Failed to create petrel instance: %v", err)
	}
	as.AddFunc("echo", "blob", hollaback)
	// and now a client
	cconf := &Config{Addr: "127.0.0.1:10298"}
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
	cconf := &Config{Addr: "999.255.255.255:10298"}
	c, err := NewTCP(cconf)
	if err == nil {
		t.Errorf("Tried connecting to invalid IP but call succeeded: `%v`", c)
	}
}

func TestNewUnixFails(t *testing.T) {
	cconf := &Config{Addr: "/foo/999.255.255.255"}
	c, err := NewUnix(cconf)
	if err == nil {
		t.Errorf("Tried connecting to invalid path but call succeeded: `%v`", c)
	}
}
