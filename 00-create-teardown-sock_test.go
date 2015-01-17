package adminsock

import (
	"testing"
	"os"
)

// create and destroy an idle adminsocket instance
func TestStartStop(t *testing.T) {
	var d Dispatch
	// fail to instantiate an adminsocket by using a terrible filename
	as, err := New("zzz/zzz/zzz/zzz", d, 0, All)
	if err == nil {
		t.Error("that should have failed, but didn't")
	}
	
	// instantiate an adminsocket
	as, err = New("test00", d, 0, All)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}
	// stat it
	fi, err := os.Stat(as.s)
	if err != nil {
		t.Errorf("Couldn't stat socket: %v", err)
	}
	fm := fi.Mode()
	if fm&os.ModeSocket == 1 {
		t.Errorf("'Socket' is not a socket %v", fm)
	}
	as.Quit()
}
