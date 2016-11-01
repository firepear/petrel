package petrel

// Copyright (c) 2014-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"fmt"
)

// Message levels control which messages will be sent to h.Msgr
const (
	All = iota
	Conn
	Error
	Fatal
)

var (
	perrs = map[string]*Perr{
		"connect": &Perr{
			100,
			Conn,
			"client connected",
			nil },
		"dispatch": &Perr{
			101,
			All,
			"dispatching",
			nil },
		"netreaderr": &Perr{
			196,
			Conn,
			"network read error",
			nil },
		"netwriteerr": &Perr{
			197,
			Conn,
			"network write error",
			nil },
		"disconnect": &Perr{
			198,
			Conn,
			"client disconnected",
			nil },
		"quit": &Perr{
			199,
			All,
			"Quit called: closing listener socket",
			nil },
		"success": &Perr{
			200,
			All,
			"reply sent",
			nil },
		"badreq": &Perr{
			400,
			All,
			"bad command",
			[]byte("PERRPERR400") },
		"nilreq": &Perr{
			401,
			All,
			"nil request",
			[]byte("PERRPERR401") },
		"plenex": &Perr{
			402,
			Error,
			"payload size limit exceeded; closing conn",
			[]byte("PERRPERR402") },
		"reqerr": &Perr{
			500,
			Error,
			"request failed",
			[]byte("PERRPERR500") },
		"internalerr": &Perr{
			501,
			Error,
			"internal error",
			nil },
		"badmac": &Perr{
			502,
			Error,
			"HMAC verification failed; closing conn",
			[]byte("PERRPERR502") },
		"listenerfail": &Perr{
			599,
			Fatal,
			"read from listener socket failed",
			nil },
	}
	perrmap = map[int]string{
		100: "connect",
		101: "dispatch",
		196: "netreaderr",
		197: "netwriteerr",
		198: "disconnect",
		199: "quit",
		200: "success",
		400: "badreq",
		401: "nilreq",
		402: "plenex",
		403: "badmac",
		500: "reqerr",
		501: "internalerr",
		599: "listenerfail" }
)

// Perr is a Petrel error -- though perhaps a better name would have
// been Pstatus. The data which is used to generate internal and
// external informational and error messages are stored as Perrs.
type Perr struct {
	Code int
	Lvl  int
	Txt  string
	xmit []byte
}

// Error implements the error interface for Perr.
func (p Perr) Error() string {
	return fmt.Sprintf("%s (%d)", p.Txt, p.Code)
}
