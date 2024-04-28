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

// Conn is a network connection, plus associated per-connection data
type Conn struct {
	nc net.Conn
	// message header buffer
	hb make([]byte, 10)
	// network read buffer
	rb make([]byte, 128)
	// TODO keep moving stuff into here
}

// ConnRead reads a message from a connection.
func ConnRead(c Conn, timeout time.Duration, plimit uint32, key []byte, seq *uint32) ([]byte, []byte, string, string, error) {
	// buffer 2: data accumulates here; payload pulled from here when done
	var b2 []byte
	// request holds the decoded request
	var request []byte
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
			return nil, nil, "disconnect", "", err
		}
		return nil, nil, "netreaderr", "no xmission header", err
	}
	if n != cap(b0) {
		return nil, nil, "netreaderr", "short read on xmission header", err
	}
	// decode the sequence id
	buf := bytes.NewReader(b0[0:4])
	err = binary.Read(buf, binary.LittleEndian, seq)
	if err != nil {
		return nil, nil, "internalerr", "could not decode seqnum", err
	}
	// decode and validate the version
	buf = bytes.NewReader(b0[4:5])
	err = binary.Read(buf, binary.LittleEndian, &pver)
	if err != nil {
		return nil, nil, "internalerr", "could not decode protocol version", err
	}
	if pver != Proto {
		return nil, nil, "internalerr", "protocol mismatch", err
	}
	// decode the request length
	buf = bytes.NewReader(b0[5:6])
	err = binary.Read(buf, binary.LittleEndian, &rlen)
	if err != nil {
		return nil, nil, "internalerr", "could not decode request length", err
	}
	// decode the payload length
	buf = bytes.NewReader(b0[6:10])
	err = binary.Read(buf, binary.LittleEndian, &plen)
	if err != nil {
		return nil, nil, "internalerr", "could not decode payload length", err
	}

	// read and decode the request
	request = make([]byte, rlen)
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
	//buf = bytes.NewReader(request)
	//err = binary.Read(buf, binary.LittleEndian, &request)
	if err != nil {
		return nil, nil, "internalerr", "could not decode request", err
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
