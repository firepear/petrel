// Copyright (c) 2014-2025 Shawn Boyette <shawn@firepear.net>. All
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

// Resp is a packaged response, recieved from a Conn
type Resp struct {
	Status  uint16
	Req     string
	Payload []byte
}

// Conn is a network connection plus associated per-connection data.
type Conn struct {
	// id; formerly cn (connection number). ignored for clients
	id uint32
	// Message sequence counter
	seq uint32
	// transmission header buffer
	hb []byte
	// payload length
	plen uint32
	// pmac stores the HMAC256
	pmac []byte
	// net.Conn, like it says on the tin
	NC net.Conn
	// Message level
	ML int
	// Response struct
	Resp Resp
	// Payload length limit
	Plim uint32
	// Network timeout
	Timeout time.Duration
	// HMAC key
	Hkey []byte
	// Msg channel
	Msgr chan *Msg
}

// ConnRead reads a transmission from a connection.
func ConnRead(c *Conn) (error) {
	// read the transmission header
	c.hb = make([]byte, 11)
	if c.Timeout > 0 {
		c.NC.SetReadDeadline(time.Now().Add(c.Timeout))
	}
	n, err := c.NC.Read(c.hb)
	if err != nil {
		if err == io.EOF {
			c.Resp.Status = 198 // (probably) clean disconnect
			return err
		}
		c.Resp.Status = 498 // read err
		return fmt.Errorf("%s: no xmission header: %v", Stats[498].Txt, err)
	}
	if n != cap(c.hb) {
		c.Resp.Status = 498 // read err
		return fmt.Errorf("%s: short read on xmission header", Stats[498].Txt)
	}

	// get data from header
	// status
	c.Resp.Status = binary.LittleEndian.Uint16(c.hb[0:1])
	// sequence id
	c.seq = binary.LittleEndian.Uint32(c.hb[2:6])
	// payload length
	c.plen = binary.LittleEndian.Uint32(c.hb[7:])
	// which cannot be greater than the payload length limit. we
	// check this again while reading the payload, because we
	// don't trust blindly
	if c.plen > c.Plim {
		c.Resp.Status = 402 // declared payload over lemgth limit
		return fmt.Errorf("%s: %d > %d", Stats[402].Txt, c.plen, c.Plim)
	}

	// read and decode the request
	req := make([]byte, c.hb[6])
	n, err = c.NC.Read(req)
	if err != nil {
		if err == io.EOF {
			c.Resp.Status = 198 // (probably) clean disconnect
			return err
		}
		c.Resp.Status = 498 // read err
		return fmt.Errorf("%s: couldn't read request: %v", Stats[498].Txt, err)
	}
	if n != cap(c.hb) {
		c.Resp.Status = 498 // read err
		return fmt.Errorf("%s: short read on request", Stats[498].Txt)
	}
	c.Resp.Req = string(req)

	// setup to read payload
	// network read buffer
	b1 := make([]byte, 128)
	// transmission accumulation buffer
	b2 := make([]byte, c.plen)
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
				c.Resp.Status = 198
				return err
			}
			c.Resp.Status = 498 // read err
			return fmt.Errorf("%s: failed to read req from socket: %v", Stats[489], err)
		}
		bread += uint32(n)
		if c.Plim > 0 && bread > c.Plim {
			c.Resp.Status = 402 // (actual) payload over length limit
			return fmt.Errorf("%s", Stats[402])
		}
		// it's easier to append everything, every time.
		// overrun is handled as soon as we stop reading
		b2 = append(b2, b1...)
	}
	// truncate payload accumulator at payload length and store as
	// the response payload
	c.Resp.Payload = b2[:c.plen]

	// finally, if we have a MAC, read and verify it
	if c.Hkey != nil {
		if c.Timeout > 0 {
			c.NC.SetReadDeadline(time.Now().Add(c.Timeout))
		}
		n, err = c.NC.Read(c.pmac)
		if err != nil {
			if err == io.EOF {
				c.Resp.Status = 198 // (probably) clean disconnect
				return err
			}
			c.Resp.Status = 498 // read err
			return fmt.Errorf("%s: failed to read HMAC from socket: %v", Stats[489], err)
		}
		mac := hmac.New(sha256.New, c.Hkey)
		mac.Write(b2)
		computedMAC := make([]byte, 44)
		base64.StdEncoding.Encode(computedMAC, mac.Sum(nil))
		if !hmac.Equal(c.pmac, computedMAC) {
			c.Resp.Status = 502 // hmac failure
			return fmt.Errorf("%v", Stats[502])
		}
	}
	return err
}

// ConnWrite writes a message to a connection.
func ConnWrite(c *Conn, request, payload []byte) error {
	if c.Timeout > 0 {
		c.NC.SetReadDeadline(time.Now().Add(c.Timeout))
	}
	_, err := c.NC.Write(marshalXmission(c, request, payload))
	if err != nil {
		// overloading response, but eh
		c.Resp.Status = 499 // write error
	}
	return err
}

// marshalXmission marshals a Msg payload into a wire-formatted
// transmission.
func marshalXmission(c *Conn, request, payload []byte) []byte {
	xmission := make([]byte, 11)
	// status
	binary.LittleEndian.PutUint16(xmission[0:], c.Resp.Status)
	// seq
	binary.LittleEndian.PutUint32(xmission[2:], c.seq)
	// encode request length
	xmission[6] = uint8(len(request))
	// encode payload length and payload
	binary.LittleEndian.PutUint32(xmission[7:], uint32(len(payload)))
	xmission = append(xmission, payload...)
	// handle HMAC if needed
	if c.Hkey != nil {
		mac := hmac.New(sha256.New, c.Hkey)
		mac.Write(payload)
		macb64 := make([]byte, 44)
		base64.StdEncoding.Encode(macb64, mac.Sum(nil))
		xmission = append(xmission, macb64...)
	}
	return xmission
}
