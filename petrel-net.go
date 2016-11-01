package petrel

// Copyright (c) 2014-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
//	"log"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"net"
	"time"
)

var (
	pverbuf = new(bytes.Buffer)
)

func init() {
	// pre-compute the binary encoding of Protover
	binary.Write(pverbuf, binary.BigEndian, Protover)
}

func connRead(c net.Conn, timeout time.Duration, plimit uint32, key []byte, seq *uint32) ([]byte, string, string, error) {
	// buffer 0 holds the transmission header
	b0 := make([]byte, 9)
	// buffer 1: network reads go here, 128B at a time
	b1 := make([]byte, 128)
	// buffer 2: data accumulates here; payload pulled from here when done
	var b2 []byte
	// pmac is the HMAC256 value which came in with the payload
	pmac := make([]byte, 32)
	// pver holds the protocol version
	var pver uint8
	// plen holds the payload length
	var plen uint32
	// bread is bytes read so far
	var bread uint32

	// read the transmission header
	if key != nil {
		// if we have an HMAC, header is 41 bytes instead of 9
		b0 = make([]byte, 41)
	}
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	n, err := c.Read(b0)
	//log.Println(n)
	//log.Println(b0)
	//log.Println(len(b0))
	//log.Println(cap(b0))
	//log.Println(key)
	//log.Println(key == nil)
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
	err = binary.Read(buf, binary.BigEndian, seq)
	if err != nil {
		return nil, "internalerr", "could not decode seqnum", err
	}
	// decode the payload length
	buf = bytes.NewReader(b0[4:8])
	err = binary.Read(buf, binary.BigEndian, &plen)
	if err != nil {
		return nil, "internalerr", "could not decode payload length", err
	}
	// decode and validate the version
	buf = bytes.NewReader(b0[8:9])
	err = binary.Read(buf, binary.BigEndian, &pver)
	if err != nil {
		return nil, "internalerr", "could not decode protocol version", err
	}
	if pver != Protover {
		return nil, "internalerr", "protocol mismatch", err
	}
	// and, optionally, extract the HMAC
	if key != nil {
		pmac = b0[9:]
		if len(pmac) != 32 {
			return nil, "netreaderr", "short read on HMAC", err
		}
	}
	//log.Println(*seq)
	//log.Println(plen)
	//log.Println(pver)
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
		if n == 0 {
			// short-circuit just in case this ever manages to happen
			return b2[:plen], "", "", err
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
		expectedMAC := mac.Sum(nil)
		//log.Println(b2)
		//log.Println(pmac)
		//log.Println(expectedMAC)
		if ! hmac.Equal(pmac, expectedMAC) {
			return nil, "badmac", "", nil
		}
	}
	return b2, "", "", err
}

func connWrite(c net.Conn, payload, key []byte, timeout time.Duration, seq uint32) (string, error) {
	xmission := []byte{}

	// encode xmit seq
	seqbuf := new(bytes.Buffer)
	err := binary.Write(seqbuf, binary.BigEndian, seq)
	if err != nil {
		return "internalerr", err
	}
	// encode payload length
	plen := new(bytes.Buffer)
	err = binary.Write(plen, binary.BigEndian, uint32(len(payload)))
	if err != nil {
		return "internalerr", err
	}
	// assemble xmission
	xmission = append(xmission, seqbuf.Bytes()...)
	xmission = append(xmission, plen.Bytes()...)
	xmission = append(xmission, pverbuf.Bytes()...)
	// encode and append HMAC if needed
	if key != nil {
		mac := hmac.New(sha256.New, key)
		mac.Write(payload)
		xmission = append(xmission, mac.Sum(nil)...)
	}

	// write to network
	xmission = append(xmission, payload...)
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	_, err = c.Write(xmission)
	if err != nil {
		return "netwriteerr", err
	}
	return "", err
}
