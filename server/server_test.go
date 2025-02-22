package server

import (
	//"errors"
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
	s, err := New(&Config{Sockname: "localhost:22", Msglvl: "debug"})
	if s != nil {
		t.Errorf("%s: should not have gotten a server, but did", t.Name())
	}
	if err == nil {
		t.Errorf("%s: err == nil on failure", t.Name())
	}
}

// handle a client connect/disconnect
func TestServerClientConnect(t *testing.T) {
	s, _ := New(&Config{Sockname: "localhost:60606", Msglvl: "debug"})
	go miniclient("localhost:60606", t)
	// TODO after connlist population is in place, this is where that testing goes
	time.Sleep(100 * time.Millisecond)
	s.Quit()
}

// we need a fake client in order to test here. but it can be really,
// really fake. we're not even going to test send/recv yet.
func miniclient(sn string, t *testing.T) {
	cc, err := pc.New(&pc.Config{Addr: sn})
	if err != nil {
		t.Errorf("%s: couldn't create client: %s", t.Name(), err)
	}
	time.Sleep(100 * time.Millisecond)
	cc.Quit()
}
