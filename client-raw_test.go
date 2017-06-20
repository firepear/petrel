package petrel

import (
	"testing"
)

func TestRawClientNewTCP(t *testing.T) {
	// instantiate TCP petrel
	asconf := &ServerConfig{Sockname: "127.0.0.1:10298", Msglvl: Fatal}
	as, err := TCPServer(asconf)
	if err != nil {
		t.Errorf("Failed to create petrel instance: %v", err)
	}
	as.Register("echo", "blob", hollaback)
	// and now a client
	cconf := &ClientConfig{Addr: "127.0.0.1:10298"}
	c, err := TCPClient(cconf)
	if err != nil {
		t.Errorf("Failed to create client: %v", err)
	}
	// and send a message
	messages := [][]byte{
		[]byte("echo just the one test"),
		[]byte("echo abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"),
	}

	for mseq, m := range messages {
		c.Seq++
		xmission, interr, err := marshalXmission(m, nil, c.Seq)
		if err != nil {
			t.Errorf("marshalXmission returned error: %s - %s", interr, err)
		}
		resp, err := c.DispatchRaw(xmission)
		if err != nil {
			t.Errorf("Dispatch returned error: %v", err)
		}
		// check sequence
		xseq := []byte{uint8(mseq + 1), 0, 0, 0}
		rseq := resp[0:4]
		for i := 0; i < 4; i++ {
			if rseq[i] != xseq[i] {
				t.Errorf("Byte %d of received seq should be %d but is %d", i, xseq[i], rseq[i])
			}
		}
		// check plen
		xplen := []byte{uint8(len(m) - 5), 0, 0, 0}
		rplen := resp[4:8]
		for i := 0; i < 4; i++ {
			if rplen[i] != xplen[i] {
				t.Errorf("Byte %d of received plen should be %d but is %d", i, xplen[i], rplen[i])
			}
		}
		// check pver
		if resp[8] != Protover {
			t.Errorf("Received protover should be %d but is %d", Protover, resp[8])
		}
		// check payload
		if string(resp[9:]) != string([]byte(m)[5:]) {
			t.Errorf("Expected `%s` but got: `%v`", m[5:], string(resp))
		}
	}
	c.Quit()
	as.Quit()
}
