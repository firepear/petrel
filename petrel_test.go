package petrel

import (
	"fmt"
	"testing"
)

// start and stop an idle petrel server
func TestMsgError(t *testing.T) {
	msg := &Msg{Cid: "foo", Seq: 8, Req: "bar", Code: 101, Txt: "baz", Err: nil}
	mstr := msg.Error()
	if mstr != "c:foo r:8 (bar, 101, in dispatch) baz" {
		t.Errorf("%s: mstr doesn't match: %s", t.Name(), mstr)
	}
	msg = &Msg{Cid: "foo", Seq: 8, Req: "bar", Code: 101, Txt: "baz",
		Err: fmt.Errorf("%d %s", 999, "quux")}
	mstr = msg.Error()
	if mstr != "c:foo r:8 (bar, 101, in dispatch) baz : 999 quux" {
		t.Errorf("%s: mstr doesn't match: %s", t.Name(), mstr)
	}

	msg = &Msg{Cid: "foo", Seq: 8, Req: "bar", Code: 2222, Txt: "baz", Err: nil}
	mstr = msg.Error()
	if mstr != "c:foo r:8 (bar, 2222) baz" {
		t.Errorf("%s: mstr doesn't match: %s", t.Name(), mstr)
	}
	msg = &Msg{Cid: "foo", Seq: 8, Req: "bar", Code: 2222, Txt: "baz",
		Err: fmt.Errorf("%d %s", 999, "quux")}
	mstr = msg.Error()
	if mstr != "c:foo r:8 (bar, 2222) baz : 999 quux" {
		t.Errorf("%s: mstr doesn't match: %s", t.Name(), mstr)
	}
}
