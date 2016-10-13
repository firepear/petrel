package petrel

import (
	"testing"
)

func hollaback(args [][]byte) ([]byte, error) {
	return args[0], nil
}

func TestClientNewUnix(t *testing.T) {
	// instantiate unix petrel
	asconf := &ServerConfig{Sockname: "/tmp/clienttest.sock", Msglvl: Fatal}
	as, err := UnixServ(asconf, 700)
	if err != nil {
		t.Errorf("Failed to create petrel instance: %v", err)
	}
	err = as.Register("echo", "blob", hollaback)
	if err != nil {
		t.Errorf("Failed to add func: %v", err)
	}
	// and now a client
	cconf := &ClientConfig{Addr: "/tmp/clienttest.sock"}
	c, err := UnixClient(cconf)
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

func TestClientNewTCP(t *testing.T) {
	// instantiate TCP petrel
	asconf := &ServerConfig{Sockname: "127.0.0.1:10298", Msglvl: Fatal}
	as, err := TCPServ(asconf)
	if err != nil {
		t.Errorf("Failed to create petrel instance: %v", err)
	}
	as.Register("echo", "blob", hollaback)
	// and now a client
	cconf := &ClientConfig{Addr: "127.0.0.1:10298"}
	c, err := TCPClient(cconf)
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

func TestClientNewTCPFails(t *testing.T) {
	cconf := &ClientConfig{Addr: "999.255.255.255:10298"}
	c, err := TCPClient(cconf)
	if err == nil {
		t.Errorf("Tried connecting to invalid IP but call succeeded: `%v`", c)
	}
}

func TestClientNewUnixFails(t *testing.T) {
	cconf := &ClientConfig{Addr: "/foo/999.255.255.255"}
	c, err := UnixClient(cconf)
	if err == nil {
		t.Errorf("Tried connecting to invalid path but call succeeded: `%v`", c)
	}
}

func TestClientClientPetrelErrs(t *testing.T) {
	// instantiate TCP petrel
	asconf := &ServerConfig{Sockname: "127.0.0.1:10298", Msglvl: Fatal}
	as, err := TCPServ(asconf)
	if err != nil {
		t.Errorf("Failed to create petrel instance: %v", err)
	}
	as.Register("echo", "blob", hollaback)
	// and now a client
	cconf := &ClientConfig{Addr: "127.0.0.1:10298"}
	c, err := TCPClient(cconf)
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
		if len(resp) != 1 && resp[0] != 255 {
			t.Errorf("len resp should 1 & resp[0] should be 255, but got len %d and '%v'", len(resp), string(resp))
		}
		if err.(*Perr).Code != perrs["badreq"].Code {
			t.Errorf("err.Code should be %d but is %v", perrs["badreq"].Code, err.(*Perr).Code)
		}
		if err.(*Perr).Txt != perrs["badreq"].Txt {
			t.Errorf("err.Txt should be %s but is %v", perrs["badreq"].Txt, err.(*Perr).Txt)
		}
		if err.Error() != "bad command (400)" {
			t.Errorf("Expected 'bad command (400)' but got '%s'", err)
		}
	}
	c.Quit()
	as.Quit()
}

