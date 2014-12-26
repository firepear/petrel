package adminsock

import (
	"testing"
	"os"
	"strings"
)

// See if our socket name builder works as expected
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

// create and destroy an idle adminsocket instance
func TestStartStop(t *testing.T) {
	var d Dispatch
	sn := buildSockName()
	// instantiate an adminsocket
	as, err := New(d, 0)
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
	as.Quit()
}
