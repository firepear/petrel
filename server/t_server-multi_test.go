package server

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	pc "github.com/firepear/petrel/client"
)

func TestServMultiServer(t *testing.T) {
	// implement an echo server
	c := &Config{Sockname: "/tmp/test03.sock", Msglvl: "conn"}
	as, err := UnixServer(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.Register("echo", echo)

	// launch clients
	rand.Seed(time.Now().Unix())
	x := 5
	for i := 0; i < x; i++ {
		go multiclient(as.s, t, i)
	}
	// wait for all clients to finish
	j := 0
	for i := 0; i < x; i++ {
		for {
			msg := <-as.Msgr
			j++
			if strings.Contains(msg.Txt, "disconnected") {
				break
			}
		}
	}
	// setting message level to Conn in New() should have resulted in
	// us seeing 10 messages instead of about 250
	if j != 10 {
		t.Errorf("Expected to see 10 Msgs but saw %v", j)
	}
	// shut down petrel
	as.Quit()
}

// connect and send 50 messages, separated by small random sleeps
func multiclient(sn string, t *testing.T, cnum int) {
	ac, err := pc.UnixClient(&pc.Config{Addr: sn})
	if err != nil {
		t.Fatalf("client instantiation failed! %s", err)
	}
	defer ac.Quit()

	for i := 0; i < 50; i++ {
		msg := fmt.Sprintf("echo message %d (which should be longer than 128 bytes to exercise a path) Lorem ipsum dolor sit amet, consectetur adipiscing elit posuere.", i)
		rmsg := fmt.Sprintf("message %d (which should be longer than 128 bytes to exercise a path) Lorem ipsum dolor sit amet, consectetur adipiscing elit posuere.", i)
		resp, err := ac.Dispatch([]byte(msg))
		if err != nil {
			t.Errorf("Error on read: %v", err)
		}
		if string(resp) != rmsg {
			t.Errorf("Expected '%v' but got '%v'", rmsg, string(resp))
		}
		time.Sleep(time.Duration(rand.Intn(25)) * time.Millisecond)
	}
}

/*
import (
	"strings"
	"testing"
)

// functions echo() and readConn() are defined in test 02. multiclient
// is defined in test 03.

func TestServMultiShutdown(t *testing.T) {
	// implement an echo server
	c := &Config{Sockname: "/tmp/test04.sock", Msglvl: "debug"}
	as, err := UnixServer(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.Register("echo", "argv", echo)

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
