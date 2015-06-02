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
	as, err := NewUnix(c, d, 700)
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
	c = Config{Sockname: "/tmp/test11.sock", Msglvl: Conn, EOM: "p!p"}
	as, err = NewUnix(c, d, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch echoclient. we should get a message about the
	// connection.
	go eomclient(as.s, t, "p!p")
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr
	// shut down asocket
	as.Quit()
}

// test EOM conditions
func eomclient(sn string, t *testing.T, eom string) {
	conn, err := net.Dial("unix", sn)
	defer conn.Close()
	if err != nil {
		t.Errorf("Couldn't connect to %v: %v", sn, err)
	}
	// send with EOM in teh middle, but not at the end; wait to make sure it gets there
	conn.Write([]byte("echo it works!" + eom + "foo"))
	time.Sleep(25 * time.Millisecond)
	res, err := readConn(conn)
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "it works!" + eom {
		t.Errorf("Expected 'it works!EOM' but got '%v'", string(res))
	}
	// finish the partial request sent last time
	conn.Write([]byte("foo bar" + eom))
	res, err = readConn(conn)
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "Unknown command 'foofoo'. Available commands: echo " + eom {
		t.Errorf("Expected unknown command help, but got '%v'", string(res))
	}
	// now send two requests at once
	conn.Write([]byte("echo thing one" + eom + "echo thing two" + eom))
	res, err = readConn(conn)
	if err != nil {
		t.Errorf("Error on read: %v", err)
	}
	if string(res) != "thing one" + eom + "thing two" + eom {
		t.Errorf("Expected 'thing oneEOMthing twoEOM' but got '%v'", string(res))
	}
}
