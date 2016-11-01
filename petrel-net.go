package petrel

// Copyright (c) 2014-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"net"
	"time"
)

func connRead(c net.Conn, timeout time.Duration, plimit uint32, key []byte, rid *uint32) ([]byte, string, string, error) {
	// buffer 0 holds the payload length & transmission id
	b0 := make([]byte, 4)
	// buffer 1: network reads go here, 128B at a time
	b1 := make([]byte, 128)
	// buffer 2: data accumulates here; payload pulled from here when done
	var b2 []byte
	// pmac is the HMAC256 value which came in with the payload
	pmac := make([]byte, 32)
	// plen holds the payload length
	var plen uint32
	// bread is bytes read so far
	var bread uint32

	// get the transmission seq id
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	n, err := c.Read(b0)
	if err != nil {
		if err == io.EOF {
			return nil, "disconnect", "", err
		}
		return nil, "netreaderr", "no xmission sequence", err
	}
	if  n != 4 {
		return nil, "netreaderr", "short read on xmission sequence", err
	}
	buf := bytes.NewReader(b0)
	err = binary.Read(buf, binary.BigEndian, rid)
	if err != nil {
		return nil, "internalerr", "could not decode xmission sequence", err
	}

	// read HMAC if we're expecting one
	if key != nil {
		if timeout > 0 {
			c.SetReadDeadline(time.Now().Add(timeout))
		}
		n, err := c.Read(pmac)
		if err != nil {
			if err == io.EOF {
				return nil, "disconnect", "", err
			}
			return nil, "netreaderr", "no HMAC", err
		}
		if  n != 32 {
			return nil, "netreaderr", "short read on HMAC", err
		}
	}

	// get the payload length
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	n, err = c.Read(b0)
	if err != nil {
		if err == io.EOF {
			return nil, "disconnect", "", err
		}
		return nil, "netreaderr", "no payload length", err
	}
	if  n != 4 {
		return nil, "netreaderr", "short read on payload length", err
	}
	buf = bytes.NewReader(b0)
	err = binary.Read(buf, binary.BigEndian, &plen)
	if err != nil {
		return nil, "internalerr", "could not decode payload length", err
	}

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

	// finally, if we have a MAC, verify it
	if key != nil {
		mac := hmac.New(sha256.New, key)
		mac.Write(b2)
		expectedMAC := mac.Sum(nil)
		if ! hmac.Equal(pmac, expectedMAC) {
			return nil, "badmac", "", nil
		}
	}
	return b2[:plen], "", "", err
}

func connWrite(c net.Conn, resp, key []byte, timeout time.Duration, rid uint32) (string, error) {
	// prepend xmission length
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, uint32(len(resp)))
	if err != nil {
		return "internalerr", err
	}
	resp2 := append(buf.Bytes(), resp...)

	// prepend HMAC
	if key != nil {
		mac := hmac.New(sha256.New, key)
		mac.Write(resp) // use base response
		resp2 = append(mac.Sum(nil), resp2...)
	}

	// prepend request id
	buf2 := new(bytes.Buffer)
	err = binary.Write(buf2, binary.BigEndian, rid)
	if err != nil {
		return "internalerr", err
	}
	resp2 = append(buf2.Bytes(), resp2...)

	// write to network
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	_, err = c.Write(resp2)
	if err != nil {
		return "netwriteerr", err
	}
	return "", err
}
