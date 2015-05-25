package asock

import (
	"errors"
	"testing"
)

// these tests check for asock.Msg implementing the Error interface
// properly.

func TestMsgError(t *testing.T) {
	c := Config{"/tmp/test13.sock", 0, 0, All, nil}
	as, err := NewUnix(c, nil)
	if err != nil {
		t.Errorf("Couldn't create socket: %v", err)
	}

	// first Msg: bare bones
	as.genMsg(1, 1, 200, 1, "", nil)
	m := <-as.Msgr
	s := m.Error()
	if s != "conn 1 req 1 status 200" {
		t.Errorf("Expected 'conn 1 req 1 status 200' but got '%v'", s)
	}

	// now with Msg.Txt
	as.genMsg(1, 1, 200, 1, "foo", nil)
	m = <-as.Msgr
	s = m.Error()
	if s != "conn 1 req 1 status 200 (foo)" {
		t.Errorf("Expected 'conn 1 req 1 status 200 (foo)' but got '%v'", s)
	}

	// and an error
	e := errors.New("something bad")
	as.genMsg(1, 1, 200, 1, "foo", e)
	m = <-as.Msgr
	s = m.Error()
	if s != "conn 1 req 1 status 200 (foo); err: something bad" {
		t.Errorf("Expected 'conn 1 req 1 status 200 (foo); err: something bad' but got '%v'", s)
	}

	as.Quit()
}
