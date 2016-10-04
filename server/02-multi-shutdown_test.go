package server

/*
import (
	"strings"
	"testing"
	"firepear.net/petrel"
)

// functions echo() and readConn() are defined in test 02. multiclient
// is defined in test 03.

func TestMultiShutdown(t *testing.T) {
	// implement an echo server
	c := &Config{Sockname: "/tmp/test04.sock", Msglvl: petrel.All}
	as, err := NewUnix(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.AddFunc("echo", "args", echo)

	// launch clients
	x := 3
	for i := 0; i < x; i++ {
		go multiclient(as.s, t, i)
	}
	// wait for all clients to connect
	for i := 0; i < x; i++ {
		for {
			msg := <-as.Msgr
			if strings.Contains(msg.Txt, "connected") {
				break
			}
		}
	}
	// do not wait for disconnect Msg. rely on shutdown to handle
	// things appropriately. This is actually the test in this file.
	as.Quit()
}
*/