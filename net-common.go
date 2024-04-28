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

// Conn is a network connection plus associated per-connection data.
type Conn struct {
	nc net.Conn
	// message sequence counter
	seq uint32
	// conn status code
	stat uint16
	// message header buffer
	hb make([]byte, 10)
	// network read buffer
	b1 make([]byte, 128)
	// transmission accumulation buffer
	b2 []byte
	// request holds the decoded request
	req []byte
	// pmac is the HMAC256
	pmac make([]byte, 44)
	// request length
	rlen uint8
	// payload length
	plen uint32
	// bytes read so far
	bread uint32
	// payload length limit
	plim uint32
}

// ConnRead reads a message from a connection.
func ConnRead(c Conn, timeout time.Duration, key []byte) ([]byte, []byte, string, string, error) {
	// read the transmission header
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	n, err := c.Read(c.hb)
	if err != nil {
		if err == io.EOF {
			return nil, nil, "disconnect", "", err
		}
		return nil, nil, "netreaderr", "no xmission header", err
	}
	if n != cap(c.hb) {
		return nil, nil, "netreaderr", "short read on xmission header", err
	}

	// get data from header
	// status
	c.stat, n = binary.LittleEndian.Uint16(c.hb[0:])
	if n != 4 {
		return nil, nil, "internalerr", "short read on status", err
	}
	// sequence id
	c.seq, n = binary.LittleEndian.Uint32(c.hb[2:])
	if n != 4 {
		return nil, nil, "internalerr", "short read on sequence", err
	}
	// request length
	c.rlen = c.hb[6]
	// payload length
	c.plen, n = binary.LittleEndian.Uint32(c.hb[7:])
	if n != 4 {
		return nil, nil, "internalerr", "short read on payloadlength", err
	}
	// which cannot be greater than the payload length limit (we
	// check this again while reading the payload, because we
	// don't trust blindly)
	if c.plen > c.plim {
		return nil, nil, "plenex", "", nil
	}

	// read and decode the request
	request = make([]byte, c.rlen)
	n, err = c.Read(request)
	if err != nil {
		if err == io.EOF {
			return nil, nil, "disconnect", "", err
		}
		return nil, nil, "netreaderr", "couldn't read request", err
	}
	if n != cap(request) {
		return nil, nil, "netreaderr", "short read on request", err
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
				return nil, nil, "disconnect", "", err
			}
			return nil, nil, "netreaderr", "failed to read req from socket", err
		}
		bread += uint32(n)
		if plimit > 0 && bread > plimit {
			return nil, nil, "plenex", "", nil
		}
		b2 = append(b2, b1[:n]...)
	}
	b2 = b2[:plen]

	// finally, if we have a MAC, read and verify it
	if key != nil {
		if timeout > 0 {
			c.SetReadDeadline(time.Now().Add(timeout))
		}
		n, err = c.Read(pmac)
		if err != nil {
			if err == io.EOF {
				return nil, nil, "disconnect", "", err
			}
			return nil, nil, "netreaderr", "failed to read req from socket", err
		}
		mac := hmac.New(sha256.New, key)
		mac.Write(b2)
		expectedMAC := make([]byte, 44)
		base64.StdEncoding.Encode(expectedMAC, mac.Sum(nil))
		if !hmac.Equal(pmac, expectedMAC) {
			return nil, nil, "badmac", "", nil
		}
	}
	return request, b2, "", "", err
}

// ConnWrite writes a message to a connection.
func ConnWrite(c net.Conn, request, payload, key []byte, timeout time.Duration, seq uint32) (string, error) {
	xmission, internalerr, err := marshalXmission(request, payload, key, seq)
	if err != nil {
		return internalerr, err
	}
	if timeout > 0 {
		c.SetReadDeadline(time.Now().Add(timeout))
	}
	_, err = c.Write(xmission)
	if err != nil {
		return "netwriteerr", err
	}
	return internalerr, err
}

// marshalXmission marshals a Msg payload into a wire-formatted
// transmission.
func marshalXmission(request, payload, key []byte, seq uint32) ([]byte, string, error) {
	xmission := []byte{}
	// encode xmit seq
	seqbuf := new(bytes.Buffer)
	err := binary.Write(seqbuf, binary.LittleEndian, seq)
	if err != nil {
		return nil, "internalerr", err
	}
	// encode request length
	rlen := new(bytes.Buffer)
	err = binary.Write(rlen, binary.LittleEndian, uint8(len(request)))
	// encode payload length
	plen := new(bytes.Buffer)
	err = binary.Write(plen, binary.LittleEndian, uint32(len(payload)))
	if err != nil {
		return nil, "internalerr", err
	}
	// assemble xmission
	xmission = append(xmission, seqbuf.Bytes()...)
	xmission = append(xmission, byte(Proto))
	xmission = append(xmission, rlen.Bytes()...)
	xmission = append(xmission, plen.Bytes()...)
	xmission = append(xmission, request...)
	xmission = append(xmission, payload...)
	// encode and append HMAC if needed
	if key != nil {
		mac := hmac.New(sha256.New, key)
		mac.Write(payload)
		macb64 := make([]byte, 44)
		base64.StdEncoding.Encode(macb64, mac.Sum(nil))
		xmission = append(xmission, macb64...)
	}
	return xmission, "", err
}
