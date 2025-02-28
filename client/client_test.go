package client

import (
	"fmt"
	//"log"
	//"sync"
	"strings"
	"testing"
	//"time"

	ps "github.com/firepear/petrel/server"
)

// bad connection string/no server listening
func TestClientNoServer(t *testing.T) {
	sn := "localhost:60606"

	// try to create with no server; c will be nil
	c, err := New(&Config{Addr: sn})
	if c != nil {
		t.Errorf("%s: New should have failed but didn't", t.Name())
	}
	if err == nil {
		t.Errorf("%s: err is nil", t.Name())
	}
}

// just connect and then disconnect
func TestIdleClient(t *testing.T) {
	sn := "localhost:60606"

	// stand up server
	s, err := ps.New(&ps.Config{Sockname: sn, Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: server creation fail: %s", t.Name(), err)
	}
	defer s.Quit()

	c, err := New(&Config{Addr: sn})
	if err != nil {
		t.Errorf("%s: %s", t.Name(), err)
	}
	c.Quit()
}

// mock and old server that doesn't handle PROTOCHECK
func TestClientNoProtohandler(t *testing.T) {
	sn := "localhost:60606"

	// stand up server
	s, err := ps.New(&ps.Config{Sockname: sn, Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: server creation fail: %s", t.Name(), err)
	}
	defer s.Quit()

	// then strip out its PROTOCHECK handler
	ok := s.RemoveHandler("PROTOCHECK")
	if !ok {
		t.Errorf("%s: removing PROTOCHECK failed", t.Name())
	}

	// try to connect; we should get a 400; c should be nil
	c, err := New(&Config{Addr: sn})
	if !strings.Contains(fmt.Sprintf("%s", err), "400") {
		t.Errorf("%s: err should be 400 here", t.Name())
	}
	if c != nil {
		t.Errorf("%s: c should be nil on 400", t.Name())
	}
}

// mock a server that always has a version mismatch
func TestClientProtoMismatch(t *testing.T) {
	sn := "localhost:60606"

	// stand up server
	s, err := ps.New(&ps.Config{Sockname: sn, Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: server creation fail: %s", t.Name(), err)
	}
	defer s.Quit()

	// replace PROTOCHECK handler with one that always
	// mismatches
	ok := s.RemoveHandler("PROTOCHECK")
	if !ok {
		t.Errorf("%s: removing PROTOCHECK failed", t.Name())
	}
	err = s.Register("PROTOCHECK", protoAlwaysMismatch)
	if err != nil {
		t.Errorf("%s: err registering handler: %s", t.Name(), err)
	}

	// try to connect; we should get a 497; c should be nil
	c, err := New(&Config{Addr: sn})
	if !strings.Contains(fmt.Sprintf("%s", err), "497") {
		t.Errorf("%s: err should be 497 here", t.Name())
	}
	if c != nil {
		t.Errorf("%s: c should be nil on 497", t.Name())
	}
}

// mock a server that reports a generic status issue, which the client
// does not handle as a special case
func TestClientProtoBadStatus(t *testing.T) {
	sn := "localhost:60606"

	// stand up server
	s, err := ps.New(&ps.Config{Sockname: sn, Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: server creation fail: %s", t.Name(), err)
	}
	defer s.Quit()

	// replace PROTOCHECK handler with one that always returns a
	// generic bad status (vs the ones we test for in client code)
	ok := s.RemoveHandler("PROTOCHECK")
	if !ok {
		t.Errorf("%s: removing PROTOCHECK failed", t.Name())
	}
	err = s.Register("PROTOCHECK", protoGenericNotSuccess)
	if err != nil {
		t.Errorf("%s: err registering handler: %s", t.Name(), err)
	}

	// try to connect; we should get a 500; c should be nil
	c, err := New(&Config{Addr: sn})
	if !strings.Contains(fmt.Sprintf("%s", err), "500") {
		t.Errorf("%s: err should be 500 here", t.Name())
	}
	if c != nil {
		t.Errorf("%s: c should be nil on !200", t.Name())
	}
}

// a replacement PROTOCHECK handler which always sends back a version
// mismatch error
func protoAlwaysMismatch(payload []byte) (uint16, []byte, error) {
	return 497, []byte{255}, nil
}

// a replacement PROTOCHECK handler which generates another non-200 status
func protoGenericNotSuccess(payload []byte) (uint16, []byte, error) {
	return 500, []byte{}, nil
}
