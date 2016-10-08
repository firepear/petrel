package petrel

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"time"
)

func connRead(c net.Conn, timeout time.Duration, reqlen int32) ([]byte, string, string, error) {
	// buffer 0 holds the message length
	b0 := make([]byte, 4)
	// buffer 1: network reads go here, 128B at a time
	b1 := make([]byte, 128)
	// buffer 2: data accumulates here; requests pulled from here
	var b2 []byte
	// message length
	var mlen int32
	// bytes read so far
	var bread int32

	// get the response message length
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	n, err := c.Read(b0)
	if err != nil {
		if err == io.EOF {
			return nil, "disconnect", "", err
		}
		return nil, "netreaderr", "no message length", err
	}
	if  n != 4 {
		return nil, "netreaderr", "short read on message length", err
	}
	buf := bytes.NewReader(b0)
	err = binary.Read(buf, binary.BigEndian, &mlen)
	if err != nil {
		return nil, "internalerr", "could not decode message length", err
	}

	for bread < mlen {
		// if there are less than 128 bytes remaining to read in this
		// message, resize b1 to fit. this avoids reading across a
		// message boundary.
		if x := mlen - bread; x < 128 {
			b1 = make([]byte, x)
		}
		if timeout > 0 {
			c.SetReadDeadline(time.Now().Add(timeout))
		}
		n, err = c.Read(b1)
		if err != nil {
			if err == io.EOF {
				return nil, "disconnect", "", err
			}
			return nil, "netreaderr", "failed to read req from socket", err
		}
		if n == 0 {
			// short-circuit just in case this ever manages to happen
			return b2[:mlen], "", "", err
		}
		bread += int32(n)
		if reqlen > 0 && bread > reqlen {
			return nil, "reqlen", "", nil
		}
		b2 = append(b2, b1[:n]...)
	}
	return b2[:mlen], "", "", err
}

func connWrite(c net.Conn, resp, key []byte, timeout time.Duration) (string, error) {
	// prepend message length
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, int32(len(resp)))
	if err != nil {
		return "internalerr", err
	}
	resp = append(buf.Bytes(), resp...)
	// write to network
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	_, err = c.Write(resp)
	if err != nil {
		return "netwriteerr", err
	}
	return "", err
}
