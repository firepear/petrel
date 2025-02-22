package server

import (
	//"errors"
	"fmt"
	//"net"
	//"os"
	//"strings"
	"testing"
	"time"

	pc "github.com/firepear/petrel/client"
)

// create and destroy an idle petrel server
func TestServerNew(t *testing.T) {
	s, err := New(&Config{Sockname: "localhost:60606", Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: failed: %s", t.Name(), err)
	}
	s.Quit()
}

// test a few failure modes, for coverage
func TestServerNewFails(t *testing.T) {
	// fail to create
	s, err := New(&Config{Sockname: "localhost:22", Msglvl: "debug"})
	if s != nil {
		t.Errorf("%s: should not have gotten a server, but did", t.Name())
	}
	if err == nil {
		t.Errorf("%s: err == nil on failure", t.Name())
	}

	// create, but try to add a handler twice
	s, err = New(&Config{Sockname: "localhost:60606", Msglvl: "debug"})
	err = s.Register("PROTOCHECK", fakehandler)
	if err == nil {
		t.Errorf("%s: added Handler twice successfully", t.Name())
	}
	s.Quit()
}

// handle a client connect/disconnect
func TestServerClientConnect(t *testing.T) {
	s, err := New(&Config{Sockname: "localhost:60606", Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: server.New failed: %s", t.Name(), err)
	}
	// connlist should have zero items
	if len(s.cl) != 0 {
		t.Errorf("%s: s.cl should have 0 len, has %d", t.Name(), len(s.cl))
	}
	// start a client and wait a tiny bit
	go miniclient("localhost:60606", t)
	time.Sleep(5 * time.Millisecond)
	// connlist should now have one entry!
	if len(s.cl) != 1 {
		t.Errorf("%s: s.cl should have len 1, has %d", t.Name(), len(s.cl))
	}
	// wait a bit more for disconnect and check that we're back at
	// zero conns
	time.Sleep(15 * time.Millisecond)
	if len(s.cl) != 0 {
		t.Errorf("%s: s.cl should have 0 len: %v", t.Name(), s.cl)
	}
	s.Quit()
}

// we need a fake client in order to test here. but it can be really,
// really fake. we're not even going to test send/recv yet.
func miniclient(sn string, t *testing.T) {
	cc, err := pc.New(&pc.Config{Addr: sn})
	if err != nil {
		t.Errorf("%s: couldn't create client: %s", t.Name(), err)
	}
	time.Sleep(15 * time.Millisecond)
	cc.Quit()
}

func fakehandler(r []byte) (uint16, []byte, error) {
	return 0, []byte{}, fmt.Errorf("fake")
}
