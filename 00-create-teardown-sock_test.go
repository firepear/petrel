package asock

import (
	"testing"
	"os"
)


func TestVersion(t *testing.T) {
	if Version != "0.16.1" {
		t.Errorf("Version mismatch: expected '0.16.1' but got '%v'", Version)
	}
}

// create and destroy an idle asocket instance
func TestStartStop(t *testing.T) {
	var d Dispatch

	// fail to instantiate an asocket by using a terrible filename
	c := Config{Sockname: "zzz/zzz/zzz/zzz", Msglvl: All}
	as, err := NewUnix(c, d, 700)
	if err == nil {
		t.Error("that should have failed, but didn't")
	}

	// instantiate an asocket
	c = Config{Sockname: "/tmp/test00.sock", Msglvl: All}
	as, err = NewUnix(c, d, 700)
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
