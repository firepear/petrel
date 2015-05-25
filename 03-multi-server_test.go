package asock

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"testing"
	"time"
)

// functions echo() and readConn() are defined in test 02.

func TestMultiServer(t *testing.T) {
	// implement an echo server
	d := make(Dispatch) // create Dispatch
	d["echo"] = &DispatchFunc{echo, "split"} // and put a function in it
	// instantiate an asocket
	c := Config{"/tmp/test03.sock", 0, 0, Conn, nil}
	as, err := NewUnix(c, d)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
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
	// shut down asocket
	as.Quit()
}

// connect and send 50 messages, separated by small random sleeps
func multiclient(sn string, t *testing.T) {
	conn, err := net.Dial("unix", sn)
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
		return
	}
	defer conn.Close()
	for i := 0; i < 50; i++ {
		msg  := fmt.Sprintf("echo message %d (which should be longer than 64 bytes to exercise a path)", i)
		rmsg := fmt.Sprintf("message %d (which should be longer than 64 bytes to exercise a path)", i)
		conn.Write([]byte(msg))
		res, err := readConn(conn)
		if err != nil {
			t.Errorf("Error on read: %v", err)
		}
		if string(res) != rmsg {
			t.Errorf("Expected '%v' but got '%v'", rmsg, string(res))
		}
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
	}
}

