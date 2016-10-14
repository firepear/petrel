package petrel

import (
	"net"
	"strings"
	"testing"
	"time"
)

// functions echo() and readConn() are defined in test 02. multiclient
// is defined in test 03.

// DeadUnix returns an instance of Asock whose listener socket closes after 100ms
func DeadUnix(c *ServerConfig) (*Server, error) {
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: c.Sockname, Net: "unix"})
	if err != nil {
		return nil, err
	}
	l.SetDeadline(time.Now().Add(100 * time.Millisecond))
	return commonNew(c, l), nil
}



func TestServENOLISTENER(t *testing.T) {
	// implement an echo server
	c := &ServerConfig{Sockname: "/tmp/test06-1.sock", Msglvl: All}
	as, err := DeadUnix(c)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}

	// wait 150ms (listener should be killed off in 100)
	time.Sleep(150 * time.Millisecond)
	// check Msgr. It should be ENOLISTENER.
	msg := <-as.Msgr
	if msg.Err == nil {
		t.Errorf("should have gotten an error, but got nil")
	}
	if msg.Txt != "read from listener socket failed" {
		t.Errorf("unexpected message: %v", msg.Txt)
	}
	if msg.Code != 599 {
		t.Errorf("dead listener should be code 599 but got %v", msg.Code)
	}
	as.Quit()
}

func TestServENOLISTENER2(t *testing.T) {
	// implement an echo server
	c := &ServerConfig{Sockname: "/tmp/test06-2.sock", Msglvl: All}
	as, err := DeadUnix(c)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	as.Register("echo", "argv", echo)

	// wait
	time.Sleep(150 * time.Millisecond)

	// check Msgr. It should be ENOLISTENER.
	msg := <-as.Msgr
	if msg.Err == nil {
		t.Errorf("should have gotten an error, but got nil")
	}
	if msg.Txt != "read from listener socket failed" {
		t.Errorf("unexpected message: %v", msg.Txt)
	}
	// oh no, our petrel is dead. gotta spawn a new one.
	as.Quit()
	c = &ServerConfig{Sockname: "/tmp/test06-3.sock", Msglvl: All}
	as, err = UnixServer(c, 700)
	if err != nil {
		t.Errorf("Couldn't spawn second listener: %v", err)
	}
	as.Register("echo", "argv", echo)

	// launch echoclient. we should get a message about the
	// connection.
	go echoclient(as.s, t)
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if !strings.HasPrefix(msg.Txt, "client connected") {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// wait for disconnect Msg
	msg = <-as.Msgr // discard cmd dispatch message
	msg = <-as.Msgr // discard reply sent message
	msg = <-as.Msgr // discard unknown cmd message
	msg = <-as.Msgr
	if msg.Err == nil {
		t.Errorf("connection drop should be an err, but got nil")
	}
	if msg.Txt != "client disconnected" {
		t.Errorf("unexpected msg.Txt: %v", msg.Txt)
	}
	// shut down petrel
	as.Quit()
}

