package adminsock

import (
	"testing"
	"net"
)

// we need a fake client in order to test here. but it can be really,
// really fake. we're not even going to test send/recv yet.
func fakeclient() {
}

// create an adminsocket. connect to it with a client which does
// nothing but wait 1/10 second before disconnecting. tear down
// adminsocket.
func TestConnHandler(t *testing.T) {
	var d Dispatch
	sn := buildSockName()
	// instantiate an adminsocket
	m, q, w, err := New(d, 0)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// launch fakeclient
	// wait for disconnect Msg
	// shut down adminsocket
	q <- true
	// there should be no error from sockAccept, because we caused the
	// shutdown
	w.Wait()
}
