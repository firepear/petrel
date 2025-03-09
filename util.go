// Copyright (c) 2014-2025 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

package petrel

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"
)

// Msg is the format which Petrel uses to communicate informational
// messages and errors to its host program via the s.Msgr channel.
type Msg struct {
	// Cid is the connection ID that the Msg is coming from.
	Cid string
	// Seq is the request number that resulted in the Msg.
	Seq uint32
	// Req is the request made
	Req string
	// Status is the numeric status indicator.
	Code uint16
	// Txt is free-form informational content
	Txt string
	// Err is the error (if any) passed upward as part of the Msg.
	Err error
}

// Error implements the error interface for Msg, returning a nicely
// (if blandly) formatted string containing all information present.
func (m *Msg) Error() string {
	if m.Err != nil {
		return fmt.Sprintf("c:%s r:%d (%s) [%d %s] %s",
			m.Cid, m.Seq, m.Req, m.Code, Stats[m.Code].Txt, m.Err)
	} else {
		return fmt.Sprintf("c:%s r:%d (%s) [%d %s]",
			m.Cid, m.Seq, m.Req, m.Code, Stats[m.Code].Txt)
	}
}

// GenId generates a SHA256 hash of the current Unix time, in
// nanoseconds. It then returns the hexadecimal string representation
// of this hash, and a "short" hash (the first 8 characters of the hex
// string, much as git does with commit hashes)
func GenId() (string, string) {
	h := sha256.Sum256([]byte(strconv.FormatInt(time.Now().UnixNano(), 16)))
	return fmt.Sprintf("%x", h), fmt.Sprintf("%x", h[:4])
}
