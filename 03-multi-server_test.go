package adminsock

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"
)

// functions echo() and readConn() are defined in test 02.

func TestMultiServer(t *testing.T) {
	// implement an echo server
	d := make(Dispatch) // create Dispatch
	d["echo"] = echo    // and put a function in it
	// instantiate an adminsocket
	as, err := New(d, 0)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch clients
	rand.Seed(time.Now().Unix())
	x := 5
	for i := 0; i < x; i++ {
		go multiclient(buildSockName(), t)
	}
	for i := 0; i < x; i++ {
		msg := <-as.Msgr
		if msg.Err != nil {
			t.Errorf("connection creation returned error: %v", msg.Err)
		}
	}
	// wait for disconnect Msg
	for i := 0; i < x; i++ {
		msg := <-as.Msgr
		if msg.Err == nil {
			t.Errorf("connection drop should be an err, but got nil")
		}
	}
	// shut down adminsocket
	as.Quit()
}

// connect and send 50 messages, separated by small random sleeps
func multiclient(sn string, t *testing.T) {
	conn, err := net.Dial("unix", sn)
	defer conn.Close()
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	for i := 0; i < 50; i++ {
		msg  := fmt.Sprintf("echo message %d", i)
		rmsg := fmt.Sprintf("message %d", i)
		conn.Write([]byte(msg))
		res := readConn(conn, t)
		if string(res) != rmsg {
			t.Errorf("Expected '%v' but got '%v'", rmsg, string(res))
		}
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
	}
}
