package adminsock

import (
	"testing"
	"os"
	"strings"
)

func TestBuildSockName(t *testing.T) {
	sn := buildSockName()
	a := strings.Split(sn, ".")
	if a[len(a) - 1] != "sock" {
		t.Errorf("Socket name doesn't end with '.sock': %v", sn)
	}
	b := strings.Split(sn, "/")
	if b[1] != "var" && b[1] != "tmp" {
		t.Errorf("Socket name doesn't begin with '/var' or '/tmp': %v", sn)
	}		
}

func TestStartStop(t *testing.T) {
	var d Dispatch
	sn := buildSockName()
	// instantiate an adminsocket
	q, e, err := New(d, 0)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// stat it
	fi, err := os.Stat(sn)
	if err != nil {
		t.Errorf("Couldn't stat socket: %v", err)
	}
	fm := fi.Mode()
	if fm&os.ModeSocket == 1 {
		t.Errorf("'Socket' is not a socket %v", fm)
	}
	// shut down adminsocket
	q <- true
	// there should be no error from sockAccept, because we caused the
	// shutdown
	more := true
	for more == true {
		select {
		case _, more = <-e:
		default:
		}
	}
	// TODO this wait needs to be de-clunkified. Come up with a good
	// solution for users. Waitgroup?
}
