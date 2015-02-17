package asock

import (
	"testing"
	"os"
)


func TestVersion(t *testing.T) {
	if Version != "0.10.0" {
		t.Errorf("Version mismatch: expected '0.10.0' but got '%v'", Version)
	}
}

// create and destroy an idle asocket instance
func TestStartStop(t *testing.T) {
	c := Config{"zzz/zzz/zzz/zzz", 0, All}
	var d Dispatch
	// fail to instantiate an asocket by using a terrible filename
	as, err := NewUnix(c, d)
	if err == nil {
		t.Error("that should have failed, but didn't")
	}
	
	// instantiate an asocket
	c = Config{"/tmp/test00.sock", 0, All}
	as, err = NewUnix(c, d)
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
