package client

import (
	"fmt"
	//"log"
	//"sync"
	"strings"
	"testing"
	//"testing/synctest"
	//"time"

	ps "github.com/firepear/petrel/server"
)

// test a few failure modes, for coverage
func TestClientNewFails(t *testing.T) {
	sn := "localhost:60606"

	// try to create with no server
	c, err := New(&Config{Addr: sn})
	if c != nil {
		t.Errorf("%s: New should have failed but didn't", t.Name())
	}
	if err == nil {
		t.Errorf("%s: err is nil", t.Name())
		c.Quit()
	}

	// stand up server for remaining tests
	s, err := ps.New(&ps.Config{Sockname: sn, Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: server creation fail: %s", t.Name(), err)
	}
	// then strip out its PROTOCHECK handler
	ok := s.RemoveHandler("PROTOCHECK")
	if !ok {
		t.Errorf("%s: removing PROTOCHECK failed", t.Name())
	}

	// try to connect; we should get a 400
	c, err = New(&Config{Addr: sn})
	if !strings.Contains(fmt.Sprintf("%s", err), "400") {
		t.Errorf("%s: err should be 400 here", t.Name())
	}
	if c != nil {
		t.Errorf("%s: c should be nil on 400", t.Name())
		c.Quit()
	}

	// shutdown the server last. doing this earlier via defer()
	// causes lockups
	//time.Sleep(15 * time.Millisecond)
	s.Quit()
}
