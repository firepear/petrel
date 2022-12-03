// Copyright (c) 2014-2022 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

package petrel

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net"
	"time"
)

const (
	// Proto is the version of the wire protocol implemented by
	// this library
	Proto = uint8(0)
)

var (
	pverbuf = new(bytes.Buffer)
)

func init() {
	// pre-compute the binary encoding of Proto
	binary.Write(pverbuf, binary.LittleEndian, Proto)
}

// ConnRead reads a message from a connection.
func ConnRead(c net.Conn, timeout time.Duration, plimit uint32, key []byte, seq *uint32) (string, []byte, string, string, error) {
	// buffer 0 holds the transmission header
	b0 := make([]byte, 10)
	// buffer 1: network reads go here, 128B at a time
	b1 := make([]byte, 128)
	// buffer 2: data accumulates here; payload pulled from here when done
	var b2 []byte
	// request holds teh decoded request
	var request string
	// pmac is the HMAC256
	pmac := make([]byte, 44)
	// pver holds the protocol version
	var pver uint8
	// rlen holds the request length
	var rlen uint8
	// plen holds the payload length
	var plen uint32
	// bread is bytes read so far
	var bread uint32

	// read the transmission header
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	n, err := c.Read(b0)
	if err != nil {
		if err == io.EOF {
			return "", nil, "disconnect", "", err
		}
		return "", nil, "netreaderr", "no xmission header", err
	}
	if n != cap(b0) {
		return "", nil, "netreaderr", "short read on xmission header", err
	}
	// decode the sequence id
	buf := bytes.NewReader(b0[0:4])
	err = binary.Read(buf, binary.LittleEndian, seq)
	if err != nil {
		return "", nil, "internalerr", "could not decode seqnum", err
	}
	// decode and validate the version
	buf = bytes.NewReader(b0[4:5])
	err = binary.Read(buf, binary.LittleEndian, &pver)
	if err != nil {
		return "", nil, "internalerr", "could not decode protocol version", err
	}
	if pver != Proto {
		return "", nil, "internalerr", "protocol mismatch", err
	}
	// decode the request length
	buf = bytes.NewReader(b0[5:6])
	err = binary.Read(buf, binary.LittleEndian, &rlen)
	if err != nil {
		return "", nil, "internalerr", "could not decode request length", err
	}
	// decode the payload length
	buf = bytes.NewReader(b0[6:10])
	err = binary.Read(buf, binary.LittleEndian, &plen)
	if err != nil {
		return "", nil, "internalerr", "could not decode payload length", err
	}

	// read and decode the request
	b0 = make([]byte, rlen)
	n, err := c.Read(b0)
	if err != nil {
		if err == io.EOF {
			return "", nil, "disconnect", "", err
		}
		return "", nil, "netreaderr", "couldn't read request", err
	}
	if n != cap(b0) {
		return nil, "netreaderr", "short read on request", err
	}
	buf = bytes.NewReader(b0)
	err = binary.Read(buf, binary.LittleEndian, &b2)
	if err != nil {
		return "", nil, "internalerr", "could not decode request", err
	}
	request = string(b2)

	// now read the payload
	b2 = []byte
	for bread < plen {
		// if there are less than 128 bytes remaining to read
		// in the payload, resize b1 to fit. this avoids
		// reading across a transmission boundary.
		if x := plen - bread; x < 128 {
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
			return "", nil, "netreaderr", "failed to read req from socket", err
		}
		bread += uint32(n)
		if plimit > 0 && bread > plimit {
			return "", nil, "plenex", "", nil
		}
		b2 = append(b2, b1[:n]...)
	}
	b2 = b2[:plen]

	// finally, if we have a MAC, read and verify it
	if key != nil {
		for bread < rlen + plen + 44 {
			// TODO read HMAC
		}
		mac := hmac.New(sha256.New, key)
		mac.Write(b2)
		expectedMAC := make([]byte, 44)
		base64.StdEncoding.Encode(expectedMAC, mac.Sum(nil))
		if !hmac.Equal(pmac, expectedMAC) {
			return "", nil, "badmac", "", nil
		}
	}
	return request, b2, "", "", err
}

// ConnReadRaw is only used by the Client, via DispatchRaw. As such it
// has no payload length checking.
func ConnReadRaw(c net.Conn, timeout time.Duration) (string, []byte, string, string, error) {
	// buffer 1: network reads go here, 128B at a time
	b1 := make([]byte, 128)
	// buffer 2: data accumulates here; payload pulled from here when done
	var b2 []byte
	for {
		if timeout > 0 {
			c.SetReadDeadline(time.Now().Add(timeout))
		}
		n, err := c.Read(b1)
		if err != nil {
			if err == io.EOF {
				return "", nil, "disconnect", "", err
			}
			return "", nil, "netreaderr", "failed to read req from socket", err
		}
		if n < 128 {
			b2 = append(b2, b1[:n]...)
			break
		}
		b2 = append(b2, b1...)
	}
	return b2, "", "", nil
}

// ConnWrite writes a message to a connection.
func ConnWrite(c net.Conn, payload, key []byte, timeout time.Duration, seq uint32) (string, error) {
	xmission, internalerr, err := marshalXmission(payload, key, seq)
	if err != nil {
		return internalerr, err
	}
	internalerr, err = ConnWriteRaw(c, timeout, xmission)
	return internalerr, err
}

// ConnWriteRaw is a lower-level function that handles network writes
// for ConnWrite and the client.
func ConnWriteRaw(c net.Conn, timeout time.Duration, xmission []byte) (string, error) {
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	_, err := c.Write(xmission)
	if err != nil {
		return "netwriteerr", err
	}
	return "", err
}

// marshalXmission marshals a Msg payload into a wire-formatted
// transmission.
func marshalXmission(payload, key []byte, seq uint32) ([]byte, string, error) {
	xmission := []byte{}
	// encode xmit seq
	seqbuf := new(bytes.Buffer)
	err := binary.Write(seqbuf, binary.LittleEndian, seq)
	if err != nil {
		return nil, "internalerr", err
	}
	// encode payload length
	plen := new(bytes.Buffer)
	err = binary.Write(plen, binary.LittleEndian, uint32(len(payload)))
	if err != nil {
		return nil, "internalerr", err
	}
	// assemble xmission
	xmission = append(xmission, seqbuf.Bytes()...)
	xmission = append(xmission, plen.Bytes()...)
	xmission = append(xmission, pverbuf.Bytes()...)
	// encode and append HMAC if needed
	if key != nil {
		mac := hmac.New(sha256.New, key)
		mac.Write(payload)
		macb64 := make([]byte, 44)
		base64.StdEncoding.Encode(macb64, mac.Sum(nil))
		xmission = append(xmission, macb64...)
	}
	xmission = append(xmission, payload...)
	return xmission, "", err
}
