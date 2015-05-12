package client

import (
	"testing"
	"firepear.net/asock"
)

// instantiate unix asock
func hollaback(args [][]byte) ([]byte, error) {
	return args[0], nil
}

func TestNewUnix(t *testing.T) {
	asdisp := make(asock.Dispatch)
	asdisp["echo"] = &asock.DispatchFunc{hollaback, "nosplit"}
	asconf := asock.Config{"/tmp/clienttest.sock", 0, asock.Fatal}
	as, err := asock.NewUnix(asconf, asdisp)
	if err != nil {
		t.Errorf("Failed to create asock instance: %v", err)
	}
	// now a client
	c, err := NewUnix("/tmp/clienttest.sock")
	if err != nil {
		t.Errorf("Failed to create client: %v", err)
	}
	// send a message
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
