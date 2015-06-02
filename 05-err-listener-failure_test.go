package asock

import (
	"net"
	"testing"
	"time"
)

// functions echo() and readConn() are defined in test 02. multiclient
// is defined in test 03.

// DeadUnix returns an instance of Asock whose listener socket closes after 100ms
func DeadUnix(c Config, d Dispatch) (*Asock, error) {
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: c.Sockname, Net: "unix"})
	if err != nil {
		return nil, err
	}
	l.SetDeadline(time.Now().Add(100 * time.Millisecond))
	return commonNew(c, d, l), nil
}



func TestENOLISTENER(t *testing.T) {
	// implement an echo server
	d := make(Dispatch) // create Dispatch
	d["echo"] = &DispatchFunc{echo, "split"} // and put a function in it
	// instantiate an asocket
	c := Config{Sockname: "/tmp/test06-1.sock", Msglvl: All}
	as, err := DeadUnix(c, d)
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

func TestENOLISTENER2(t *testing.T) {
	// implement an echo server
	d := make(Dispatch) // create Dispatch
	d["echo"] = &DispatchFunc{echo, "split"} // and put a function in it
	// instantiate an asocket
	c := Config{Sockname: "/tmp/test06-2.sock", Msglvl: All}
	as, err := DeadUnix(c, d)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
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
	// oh no, our asocket is dead. gotta spawn a new one.
	as.Quit()
	c = Config{Sockname: "/tmp/test06-3.sock", Msglvl: All}
	as, err = NewUnix(c, d, 700)
	if err != nil {
		t.Errorf("Couldn't spawn second listener: %v", err)
	}
	// launch echoclient. we should get a message about the
	// connection.
	go echoclient(as.s, t)
	msg = <-as.Msgr
	if msg.Err != nil {
		t.Errorf("connection creation returned error: %v", msg.Err)
	}
	if msg.Txt != "client connected" {
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
	// shut down asocket
	as.Quit()
}

