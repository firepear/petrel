package server

import (
	"errors"
	"testing"

	"firepear.net/petrel"
)

// these tests check for petrel.Msg implementing the Error interface
// properly.

func TestMsgError(t *testing.T) {
	c := &Config{Sockname: "/tmp/test13.sock", Msglvl: petrel.All}
	as, err := NewUnix(c, 700)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}

	// first Msg: bare bones
	as.genMsg(1, 1, petrel.Perrs["success"], "", nil)
	m := <-as.Msgr
	s := m.Error()
	if s != "conn 1 req 1 status 200 (reply sent)" {
		t.Errorf("Expected 'conn 1 req 1 status 200 (reply sent)' but got '%v'", s)
	}

	// now with Msg.Txt
	as.genMsg(1, 1, petrel.Perrs["success"], "foo", nil)
	m = <-as.Msgr
	s = m.Error()
	if s != "conn 1 req 1 status 200 (reply sent: [foo])" {
		t.Errorf("Expected 'conn 1 req 1 status 200 (reply sent: [foo])' but got '%v'", s)
	}

	// and an error
	e := errors.New("something bad")
	as.genMsg(1, 1, petrel.Perrs["success"], "foo", e)
	m = <-as.Msgr
	s = m.Error()
	if s != "conn 1 req 1 status 200 (reply sent: [foo]); err: something bad" {
		t.Errorf("Expected 'conn 1 req 1 status 200 (reply sent: [foo]); err: something bad' but got '%v'", s)
	}
	as.Quit()
}

