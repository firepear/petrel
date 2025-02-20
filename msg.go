package petrel

// Copyright (c) 2014-2025 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"fmt"
)

// GenMsg creates messages and sends them to the Msgr channel, if they
// match or exceed Conn.ML
func (c *Conn) GenMsg(status uint16, req string, err error) {
	// if this message's level is below the instance's level, don't
	// generate the message
	if Stats[status].Lvl < c.ML {
		return
	}
	c.Msgr <- &Msg{c.Id, c.Seq, req, status, err}
}

// Msg is the format which Petrel uses to communicate informational
// messages and errors to its host program via the s.Msgr channel.
type Msg struct {
	// Cid is the connection ID that the Msg is coming from.
	Cid uint32
	// Seq is the request number that resulted in the Msg.
	Seq uint32
	// Req is the request made
	Req string
	// Status is the numeric status indicator.
	Code uint16
	// Err is the error (if any) passed upward as part of the Msg.
	Err error
}

// Error implements the error interface for Msg, returning a nicely
// (if blandly) formatted string containing all information present.
func (m *Msg) Error() string {
	err := fmt.Sprintf("conn %d req %d (%s) %d: %s", m.Cid, m.Seq, m.Req, m.Code, Stats[m.Code].Txt)
	if m.Err != nil {
		err = err + fmt.Sprintf("; %s", m.Err)
	}
	return err
}
