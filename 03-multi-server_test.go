package adminsock

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
	d["echo"] = echo    // and put a function in it
	// instantiate an adminsocket
	as, err := New("test03", d, 0, All)
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
	for i := 0; i < x; i++ {
		for {
			msg := <-as.Msgr
			if strings.Contains(msg.Txt, "lost") {
				break
			}
		}
	}
	// shut down adminsocket
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

