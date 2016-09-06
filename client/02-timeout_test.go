package pclient

import (
	"testing"
	"firepear.net/petrel"
	"strings"
	"time"
)

func waitwhat(args [][]byte) ([]byte, error) {
	time.Sleep(40 * time.Millisecond)
	return args[0], nil
}

func TestClientTimeout(t *testing.T) {
	// instantiate unix petrel
	asconf := &petrel.Config{Sockname: "/tmp/clienttest2.sock", Msglvl: petrel.Fatal}
	as, err := petrel.NewUnix(asconf, 700)
	if err != nil {
		t.Fatalf("Failed to create petrel instance: %v", err)
	}
	as.AddFunc("echo", "blob", hollaback)
	as.AddFunc("slow", "blob", waitwhat)

	// and now a client
	cconf := &Config{Addr: "/tmp/clienttest2.sock", Timeout: 25}
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

	// now send a message which will take too long to come back
	resp, err = c.Dispatch([]byte("slow just the one test, slowly"))
	if err == nil {
		t.Errorf("Dispatch should have timed out, but no error. Got: %v", string(resp))
	}
	if err != nil && !strings.HasSuffix(err.Error(), "i/o timeout") {
		t.Errorf("Expected read timeout, but got: %v", err)
	}
	// wait a bit and see what we get if we check the socket again
	time.Sleep(40 * time.Millisecond)
	resp, err = c.Read()
	if err != nil {
		t.Errorf("Read returned error: %v", err)
	}
	if string(resp) != "just the one test, slowly" {
		t.Errorf("Expected `just the one test, slowly` but got: `%v`", string(resp))
	}
	c.Close()
	as.Quit()
}
