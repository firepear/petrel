package server

import (
	//"errors"
	"net"
	//"os"
	//"strings"
	"testing"
	"time"
	//p "github.com/firepear/petrel"
)

// create and destroy an idle petrel server
func TestServerNew(t *testing.T) {
	s, err := New(&Config{Sockname: "localhost:60606", Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: failed: %s", t.Name(), err)
	}
	s.Quit()
}

// test a few failure modes
func TestServerNewFails(t *testing.T) {
	s, err := New(&Config{Sockname: "localhost:22", Msglvl: "debug"})
	if s != nil {
		t.Errorf("%s: should not have gotten a server, but did", t.Name())
	}
	if err == nil {
		t.Errorf("%s: err == nil on failure", t.Name())
	}
}

// we need a fake client in order to test here. but it can be really,
// really fake. we're not even going to test send/recv yet.
func fakeclient(sn string, t *testing.T) {
	conn, err := net.Dial("unix", sn)
	defer conn.Close()
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	time.Sleep(100 * time.Millisecond)
}
