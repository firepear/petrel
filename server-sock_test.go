package petrel

import (
	"testing"
	"os"
	
)


// create and destroy an idle petrel instance
func TestServStartStop(t *testing.T) {
	// fail to instantiate petrel by using a terrible filename
	c := &ServerConfig{Sockname: "zzz/zzz/zzz/zzz", Msglvl: All}
	as, err := UnixServ(c, 700)
	if err == nil {
		t.Error("that should have failed, but didn't")
	}

	// instantiate petrel
	c = &ServerConfig{Sockname: "/tmp/test00.sock", Msglvl: All}
	as, err = UnixServ(c, 700)
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