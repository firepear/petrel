package asock

import (
	"net"
	"testing"
	"time"
)

// implement an echo server
func TestEOMServer(t *testing.T) {
	d := make(Dispatch) // create Dispatch
	d["echo"] = &DispatchFunc{echo, "split"} // and put a function in it
	// instantiate an asocket
	c := Config{Sockname: "/tmp/test11.sock", Msglvl: Conn}
	as, err := NewUnix(c, d)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch echoclient. we should get a message about the
	// connection.
	go eomclient(as.s, t, "\n\n")
	msg := <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr
	// shut down asocket
	as.Quit()

	// instantiate an asocket, this time with custom EOM
	/*c := Config{Sockname: "/tmp/test11.sock", Msglvl: Conn}
	as, err := NewUnix(c, d)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch echoclient. we should get a message about the
	// connection.
	go echoclient(as.s, t) */
}

// this time our (less) fake client will send a string over the
// connection and (hopefully) get it echoed back.
func eomclient(sn string, t *testing.T, eom string) {
	conn, err := net.Dial("unix", sn)
	defer conn.Close()
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	conn.Write([]byte("echo it works!" + eom + "foo"))
	time.Sleep(50 * time.Millisecond)
	res, err := readConn(conn)
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "it works!" {
		t.Errorf("Expected 'it works!' but got '%v'", string(res))
	}
	// for bonus points, let's send a bad command
	conn.Write([]byte("foo bar" + eom))
	res, err = readConn(conn)
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "Unknown command 'foofoo'\nAvailable commands:\n    echo\n" {
		t.Errorf("Expected unknown command help, but got '%v'", string(res))
	}
}
