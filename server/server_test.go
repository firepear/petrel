package server

import (
	"fmt"
	//	"log"
	"sync"
	"syscall"
	"testing"
	"time"

	pc "github.com/firepear/petrel/client"
)

var sn = "localhost:60606"

// start and stop an idle petrel server
func TestServerNew(t *testing.T) {
	s, err := New(&Config{Sockname: sn, Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: failed: %s", t.Name(), err)
	}
	s.Quit()
}

// start and halt (via signal) a server
func TestServerSig(t *testing.T) {
	s, err := New(&Config{Sockname: sn, Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: failed: %s", t.Name(), err)
	}
	s.sig <- syscall.SIGINT
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
	s, err = New(&Config{Sockname: sn, Msglvl: "debug"})
	err = s.Register("PROTOCHECK", fakehandler)
	if err == nil {
		t.Errorf("%s: added Handler twice successfully", t.Name())
	}
	s.Quit()
}

// handle a client connect/disconnect
func TestServerClientConnect(t *testing.T) {
	//synctest.Run(func() {
	s, err := New(&Config{Sockname: sn, Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: server.New failed: %s", t.Name(), err)
	}
	// connlist should have zero items
	i := lenSyncMap(&s.cl)
	if i != 0 {
		t.Errorf("%s: s.cl should have 0 len, has %d", t.Name(), i)
	}

	// start a client and wait a tiny bit
	go miniclient(sn, t)
	time.Sleep(15 * time.Millisecond)
	//synctest.Wait()

	// connlist should now have one entry!
	i = lenSyncMap(&s.cl)
	if i != 1 {
		t.Errorf("%s: s.cl should have len 1, has %d", t.Name(), i)
	}
	//synctest.Wait()

	// wait until we're back at zero conns
	for lenSyncMap(&s.cl) > 0 {
		time.Sleep(1 * time.Millisecond)
	}
	s.Quit()
	//})
}

// make sure many clients at once works properly
func TestServerClientClobber(t *testing.T) {
	s, err := New(&Config{Sockname: sn, Msglvl: "debug"})
	if err != nil {
		t.Errorf("%s: server.New failed: %s", t.Name(), err)
	}
	// launch 100 clients
	for range 100 {
		go miniclient(sn, t)
	}
	time.Sleep(5 * time.Millisecond)
	// we should have at least 25 in the list
	i := lenSyncMap(&s.cl)
	if i < 25 {
		t.Errorf("%s: s.cl should have many clients; have %d", t.Name(), i)
	}
	// test connlist len until client drops
	for lenSyncMap(&s.cl) > 0 {
		time.Sleep(2 * time.Millisecond)
	}
	s.Quit()
}

// start and stop an idle petrel server -- but with high msglvl
func TestServerHighMsglvl(t *testing.T) {
	s, err := New(&Config{Sockname: sn, Msglvl: "fatal"})
	if err != nil {
		t.Errorf("%s: failed: %s", t.Name(), err)
	}
	go miniclient(sn, t)
	// test connlist len until client drops
	for lenSyncMap(&s.cl) > 0 {
		time.Sleep(2 * time.Millisecond)
	}
	s.Quit()
}

// start a server with a very low payload length limit
func TestServerSmallPayload(t *testing.T) {
	s, err := New(&Config{Sockname: sn, Xferlim: 15})
	if err != nil {
		t.Errorf("%s: failed: %s", t.Name(), err)
	}
	s.Register("FAKE", fakehandler)
	cc, err := pc.New(&pc.Config{Addr: sn})
	if err != nil {
		t.Errorf("%s: couldn't create client: %s", t.Name(), err)
	}
	err = cc.Dispatch("FAKE", []byte("this is too many bytes for xferlimit"))
	if cc.Resp.Status != 402 {
		t.Errorf("%s: status should be 402 here: %d", t.Name(), cc.Resp.Status)
	}
	if string(cc.Resp.Payload) != "36 > 15" {
		t.Errorf("%s: unexpected response %s", t.Name(), string(cc.Resp.Payload))
	}
	cc.Quit()
	s.Quit()
}

/*////////////////////////////////////////////////////////////////////////
  // below this point are the functions used by tests
  ////////////////////////////////////////////////////////////////////////*/

// miniclient instantiates a client which does nothing for 15
// milliseconds. useful for baseline tests
func miniclient(sn string, t *testing.T) {
	cc, err := pc.New(&pc.Config{Addr: sn})
	if err != nil {
		t.Errorf("%s: couldn't create client: %s", t.Name(), err)
	}
	time.Sleep(15 * time.Millisecond)
	cc.Quit()
}

// lenSyncMap counts the items in the server's sync.Map
func lenSyncMap(m *sync.Map) int {
	var i int
	m.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	return i
}

// fakehandler is a handler that does nothing
func fakehandler(r []byte) (uint16, []byte, error) {
	return 0, []byte{}, fmt.Errorf("fake")
}
