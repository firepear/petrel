package petrel

// Copyright (c) 2014-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/base64"
	"io"
	"net"
	"time"
)

var (
	pverbuf = new(bytes.Buffer)
)

func init() {
	// pre-compute the binary encoding of Protover
	binary.Write(pverbuf, binary.LittleEndian, Protover)
}

func connRead(c net.Conn, timeout time.Duration, plimit uint32, key []byte, seq *uint32) ([]byte, string, string, error) {
	// buffer 0 holds the transmission header
	b0 := make([]byte, 9)
	// buffer 1: network reads go here, 128B at a time
	b1 := make([]byte, 128)
	// buffer 2: data accumulates here; payload pulled from here when done
	var b2 []byte
	// pmac is the HMAC256 value which came in with the payload
	pmac := make([]byte, 44)
	// pver holds the protocol version
	var pver uint8
	// plen holds the payload length
	var plen uint32
	// bread is bytes read so far
	var bread uint32

	// read the transmission header
	if key != nil {
		// if we have an HMAC, header is 53 bytes instead of 9
		b0 = make([]byte, 53)
	}
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	n, err := c.Read(b0)
	if err != nil {
		if err == io.EOF {
			return nil, "disconnect", "", err
		}
		return nil, "netreaderr", "no xmission header", err
	}
	if  n != cap(b0) {
		return nil, "netreaderr", "short read on xmission header", err
	}
	// decode the sequence id
	buf := bytes.NewReader(b0[0:4])
	err = binary.Read(buf, binary.LittleEndian, seq)
	if err != nil {
		return nil, "internalerr", "could not decode seqnum", err
	}
	// decode the payload length
	buf = bytes.NewReader(b0[4:8])
	err = binary.Read(buf, binary.LittleEndian, &plen)
	if err != nil {
		return nil, "internalerr", "could not decode payload length", err
	}
	// decode and validate the version
	buf = bytes.NewReader(b0[8:9])
	err = binary.Read(buf, binary.LittleEndian, &pver)
	if err != nil {
		return nil, "internalerr", "could not decode protocol version", err
	}
	if pver != Protover {
		return nil, "internalerr", "protocol mismatch", err
	}
	// and, optionally, extract the HMAC
	if key != nil {
		pmac = b0[9:]
		if len(pmac) != 44 {
			return nil, "netreaderr", "short read on HMAC", err
		}
	}

	// now read the payload
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
			return nil, "netreaderr", "failed to read req from socket", err
		}
		bread += uint32(n)
		if plimit > 0 && bread > plimit {
			return nil, "plenex", "", nil
		}
		b2 = append(b2, b1[:n]...)
	}
	b2 = b2[:plen]

	// finally, if we have a MAC, verify it
	if key != nil {
		mac := hmac.New(sha256.New, key)
		mac.Write(b2)
		expectedMAC := make([]byte, 44)
		base64.StdEncoding.Encode(expectedMAC, mac.Sum(nil))
		if ! hmac.Equal(pmac, expectedMAC) {
			return nil, "badmac", "", nil
		}
	}
	return b2, "", "", err
}

// connReadRaw is only used by the Client, via DispatchRaw. As such it
// has no payload length checking.
func connReadRaw(c net.Conn, timeout time.Duration) ([]byte, string, string, error) {
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
				return nil, "disconnect", "", err
			}
			return nil, "netreaderr", "failed to read req from socket", err
		}
		if n < 128 {
			b2 = append(b2, b1[:n]...)
			break
		}
		b2 = append(b2, b1...)
	}
	return b2, "", "", nil
}

func connWrite(c net.Conn, payload, key []byte, timeout time.Duration, seq uint32) (string, error) {
	xmission, internalerr, err := marshalXmission(payload, key, seq)
	if err != nil {
		return internalerr, err
	}
	internalerr, err = connWriteRaw(c, timeout, xmission)
	return internalerr, err
}

func connWriteRaw(c net.Conn, timeout time.Duration, xmission []byte) (string, error) {
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
// transmission. The format is:
//
//    Sequence        uint32 (4 bytes)
//    Payload length  uint32 (4 bytes)
//    Protocol ver    uint8  (1 byte)
//    HMAC            32 bytes, optional
//    Payload         Per payload length
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
