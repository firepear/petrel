package petrel

// Copyright (c) 2014-2022 Shawn Boyette <shawn@firepear.net>. All
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
	Errs = map[string]*Perr{
		"connect": {
			100,
			Conn,
			"client connected",
			nil},
		"dispatch": {
			101,
			All,
			"dispatching",
			nil},
		"netreaderr": {
			196,
			Conn,
			"network read error",
			nil},
		"netwriteerr": {
			197,
			Conn,
			"network write error",
			nil},
		"disconnect": {
			198,
			Conn,
			"client disconnected",
			nil},
		"quit": {
			199,
			All,
			"Quit called: closing listener socket",
			nil},
		"success": {
			200,
			All,
			"reply sent",
			nil},
		"badreq": {
			400,
			All,
			"bad command",
			[]byte("PERRPERR400")},
		"nilreq": {
			401,
			All,
			"nil request",
			[]byte("PERRPERR401")},
		"plenex": {
			402,
			Error,
			"payload size limit exceeded; closing conn",
			[]byte("PERRPERR402")},
		"reqerr": {
			500,
			Error,
			"request failed",
			[]byte("PERRPERR500")},
		"internalerr": {
			501,
			Error,
			"internal error",
			nil},
		"badmac": {
			502,
			Error,
			"HMAC verification failed; closing conn",
			[]byte("PERRPERR502")},
		"listenerfail": {
			599,
			Fatal,
			"read from listener socket failed",
			nil},
	}
	Errmap = map[int]string{
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
		599: "listenerfail"}
)

// Perr is a Petrel error -- though perhaps a better name would have
// been Pstatus. The data which is used to generate internal and
// external informational and error messages are stored as Perrs.
type Perr struct {
	Code int
	Lvl  int
	Txt  string
	Xmit []byte
}

// Error implements the error interface for Perr.
func (p Perr) Error() string {
	return fmt.Sprintf("%s (%d)", p.Txt, p.Code)
}
