// Copyright (c) 2014-2022 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

package petrel

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// Conn is a network connection plus associated per-connection data.
type Conn struct {
	NC net.Conn
	// network timeout
	Timeout time.Duration
	// message sequence counter
	Seq uint32
	// conn/req status code
	Stat uint16
	// transmission header buffer
	hb []byte
	// HMAC key
	Hkey []byte
	// pmac is the HMAC256
	pmac []byte
	// request length
	rlen uint8
	// payload length
	plen uint32
	// payload length limit
	Plim uint32
}

// ConnRead reads a transmission from a connection.
func ConnRead(c *Conn) ([]byte, []byte, error) {
	// read the transmission header
	if c.Timeout > 0 {
		c.NC.SetReadDeadline(time.Now().Add(c.Timeout))
	}
	n, err := c.NC.Read(c.hb)
	if err != nil {
		if err == io.EOF {
			// (probably) clean disconnect
			c.Stat = 198
			return nil, nil, err
		}
		c.Stat = 498
		return nil, nil, fmt.Errorf("%s: no xmission header: %v", Stats[498], err)
	}
	if n != cap(c.hb) {
		c.Stat = 498
		return nil, nil, fmt.Errorf("%s: short read on xmission header", Stats[498])
	}

	// get data from header
	// status
	c.Stat = binary.LittleEndian.Uint16(c.hb[0:])
	// sequence id
	c.Seq = binary.LittleEndian.Uint32(c.hb[2:])
	// request length
	c.rlen = c.hb[6]
	// payload length
	c.plen = binary.LittleEndian.Uint32(c.hb[7:])
	// which cannot be greater than the payload length limit. we
	// check this again while reading the payload, because we
	// don't trust blindly
	if c.plen > c.Plim {
		c.Stat = 402
		return nil, nil, fmt.Errorf("%s", Stats[402])
	}

	// read and decode the request
	req := make([]byte, c.rlen)
	n, err = c.NC.Read(req)
	if err != nil {
		if err == io.EOF {
			// (probably) clean disconnect
			c.Stat = 198
			return nil, nil, err
		}
		c.Stat = 498
		return nil, nil, fmt.Errorf("%s: couldn't read request: %v", Stats[498], err)
	}
	if n != cap(c.hb) {
		c.Stat = 498
		return nil, nil, fmt.Errorf("%s: short read on request", Stats[498])
	}

	// setup to read payload
	// network read buffer
	b1 := make([]byte, 128)
	// transmission accumulation buffer
	b2 := []byte{}
	// accumulated bytes read
	bread := uint32(0)

	// now read the payload
	for bread < c.plen {
		// if there are less than 128 bytes remaining to read
		// in the payload, resize b1 to fit. this avoids
		// reading across a transmission boundary.
		if x := c.plen - bread; x < 128 {
			b1 = make([]byte, x)
		}
		if c.Timeout > 0 {
			c.NC.SetReadDeadline(time.Now().Add(c.Timeout))
		}
		n, err = c.NC.Read(b1)
		if err != nil {
			if err == io.EOF {
				c.Stat = 198
				return nil, nil, err
			}
			c.Stat = 498
			return nil, nil, fmt.Errorf("%s: failed to read req from socket: %v", Stats[489], err)
		}
		bread += uint32(n)
		if c.Plim > 0 && bread > c.Plim {
			c.Stat = 402
			return nil, nil, fmt.Errorf("%s", Stats[402])
		}
		b2 = append(b2, b1...)
		//b2 = append(b2, b1[:n]...)
	}
	b2 = b2[:c.plen]

	// finally, if we have a MAC, read and verify it
	if c.Hkey != nil {
		if c.Timeout > 0 {
			c.NC.SetReadDeadline(time.Now().Add(c.Timeout))
		}
		n, err = c.NC.Read(c.pmac)
		if err != nil {
			if err == io.EOF {
				c.Stat = 198
				return nil, nil, err
			}
			c.Stat = 498
			return nil, nil, fmt.Errorf("%s: failed to read HMAC from socket: %v", Stats[489], err)
		}
		mac := hmac.New(sha256.New, c.Hkey)
		mac.Write(b2)
		computedMAC := make([]byte, 44)
		base64.StdEncoding.Encode(computedMAC, mac.Sum(nil))
		if !hmac.Equal(c.pmac, computedMAC) {
			c.Stat = 502
			return nil, nil, fmt.Errorf("%v", Stats[502])
		}
	}
	return req, b2, err
}

// ConnWrite writes a message to a connection.
func ConnWrite(c *Conn, request, payload []byte) error {
	xmission := marshalXmission(c, request, payload)
	if c.Timeout > 0 {
		c.NC.SetReadDeadline(time.Now().Add(c.Timeout))
	}
	_, err := c.NC.Write(xmission)
	if err != nil {
		return err
	}
	return err
}

// marshalXmission marshals a Msg payload into a wire-formatted
// transmission.
func marshalXmission(c *Conn, request, payload []byte) []byte {
	xmission := []byte{}
	// status
	binary.LittleEndian.PutUint16(xmission[0:], c.Stat)
	// seq
	binary.LittleEndian.PutUint32(xmission[2:], c.Seq)
	// encode request length
	xmission[6] = uint8(len(request))
	// encode payload length
	binary.LittleEndian.PutUint32(xmission[7:], uint32(len(payload)))
	if c.Hkey != nil {
		mac := hmac.New(sha256.New, c.Hkey)
		mac.Write(payload)
		macb64 := make([]byte, 44)
		base64.StdEncoding.Encode(macb64, mac.Sum(nil))
		xmission = append(xmission, macb64...)
	}
	return xmission
}
