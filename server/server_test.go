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
func TestServStartStop(t *testing.T) {
	s, err := New(&Config{Sockname: "localhost:60606", Msglvl: "debug"})
	if err != nil {
		t.Errorf("idle server creation test failed: %s", err)
	}
	s.Quit()
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
