package petrel

import (
	"testing"
)

func hollaback(args [][]byte) ([]byte, error) {
	return args[0], nil
}

func TestNewUnix(t *testing.T) {
	// instantiate unix petrel
	asconf := &server.Config{Sockname: "/tmp/clienttest.sock", Msglvl: Fatal}
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
	// instantiate TCP petrel
	asconf := &server.Config{Sockname: "127.0.0.1:10298", Msglvl: Fatal}
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

func TestClientPetrelErrs(t *testing.T) {
	// instantiate TCP petrel
	asconf := &server.Config{Sockname: "127.0.0.1:10298", Msglvl: Fatal}
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
	// and send a bad command
	resp, err := c.Dispatch([]byte("bad command"))
	if err == nil {
		t.Errorf("bad command should have returned an error, but got %s", string(resp))
		t.Errorf("resp len should be 11 but is %d", len(resp))
		t.Errorf("resp[0] should be 80 (P) but is %d (%s)", resp[0], string(resp[0]))
		t.Errorf("resp[0:8] should be 'PERRPERR' but is %s", string(resp[0:8]))
	} else {
		if err.Error() != "bad command (400)" {
			t.Errorf("Expected 'bad command (400)' but got '%s'", err)
		}
	}
	c.Close()
	as.Quit()
}

