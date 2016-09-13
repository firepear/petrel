package server

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"firepear.net/petrel/client"
)

func TestMultiServer(t *testing.T) {
	// implement an echo server
	c := &Config{Sockname: "/tmp/test03.sock", Msglvl: Conn}
	as, err := NewUnix(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.AddFunc("echo", "args", echo)

	// launch clients
	rand.Seed(time.Now().Unix())
	x := 5
	for i := 0; i < x; i++ {
		go multiclient(as.s, t)
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
func multiclient(sn string, t *testing.T) {
	ac, err := pclient.NewUnix(&pclient.Config{Addr: sn})
	if err != nil {
		t.Fatalf("pclient instantiation failed! %s", err)
	}
	defer ac.Close()

	for i := 0; i < 50; i++ {
		msg  := fmt.Sprintf("echo message %d (which should be longer than 128 bytes to exercise a path) Lorem ipsum dolor sit amet, consectetur adipiscing elit posuere.", i)
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
