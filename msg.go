package petrel

// Copyright (c) 2014-2025 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"fmt"
)

// genMsg creates messages and sends them to the Msgr channel.
func (c *Conn) genMsg(status uint16, err error) {
	// if this message's level is below the instance's level, don't
	// generate the message
	if Stats[status].Lvl < c.ML {
		return
	}
	c.Msgr <- &Msg{c.id, c.seq, status, err}
}

// Msg is the format which Petrel uses to communicate informational
// messages and errors to its host program via the s.Msgr channel.
type Msg struct {
	// Cid is the connection ID that the Msg is coming from.
	Cid uint32
	// Seq is the request number that resulted in the Msg.
	Seq uint32
	// Code is the numeric status indicator.
	Code uint16
	// Err is the error (if any) passed upward as part of the Msg.
	Err error
}

// Error implements the error interface for Msg, returning a nicely
// (if blandly) formatted string containing all information present.
func (m *Msg) Error() string {
	err := fmt.Sprintf("conn %d req %d status %d: %s", m.Cid, m.Seq, m.Code, Stats[m.Code].Txt)
	if m.Err != nil {
		err = err + fmt.Sprintf("; err: %s", m.Err)
	}
	return err
}
