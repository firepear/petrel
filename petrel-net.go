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

func connRead(c net.Conn, timeout time.Duration, reqlen uint32, key []byte, rid *uint32) ([]byte, string, string, error) {
	// buffer 0 holds the message length & message id
	b0 := make([]byte, 4)
	// buffer 1: network reads go here, 128B at a time
	b1 := make([]byte, 128)
	// buffer 2: data accumulates here; requests pulled from here
	var b2 []byte
	// msgmac is the HMAC256 value read from the network
	msgmac := make([]byte, 32)
	// message length
	var mlen uint32
	// bytes read so far
	var bread uint32

	// get the message seq id
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	n, err := c.Read(b0)
	if err != nil {
		if err == io.EOF {
			return nil, "disconnect", "", err
		}
		return nil, "netreaderr", "no message sequence", err
	}
	if  n != 4 {
		return nil, "netreaderr", "short read on message sequence", err
	}
	buf := bytes.NewReader(b0)
	err = binary.Read(buf, binary.BigEndian, rid)
	if err != nil {
		return nil, "internalerr", "could not decode message sequence", err
	}

	// read HMAC if we're expecting one
	if key != nil {
		if timeout > 0 {
			c.SetReadDeadline(time.Now().Add(timeout))
		}
		n, err := c.Read(msgmac)
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

	// get the response length
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	n, err = c.Read(b0)
	if err != nil {
		if err == io.EOF {
			return nil, "disconnect", "", err
		}
		return nil, "netreaderr", "no message length", err
	}
	if  n != 4 {
		return nil, "netreaderr", "short read on message length", err
	}
	buf = bytes.NewReader(b0)
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
		bread += uint32(n)
		if reqlen > 0 && bread > reqlen {
			return nil, "reqlen", "", nil
		}
		b2 = append(b2, b1[:n]...)
	}

	// finally, if we have a MAC, verify it
	if key != nil {
		mac := hmac.New(sha256.New, key)
		mac.Write(b2)
		expectedMAC := mac.Sum(nil)
		if ! hmac.Equal(msgmac, expectedMAC) {
			return nil, "badmac", "", nil
		}
	}
	return b2[:mlen], "", "", err
}

func connWrite(c net.Conn, resp, key []byte, timeout time.Duration, rid uint32) (string, error) {
	// prepend message length
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
