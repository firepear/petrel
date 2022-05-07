package petrel

// Copyright (c) 2014-2022 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

import (
	"fmt"
)

// Message levels control which messages will be sent to h.Msgr
const (
	Debug = iota
	Conn
	Error
	Fatal
)

var (
	// Errs is the map of Perr instances. It is used by Msg
	// handling code throughout the Petrel packages.
	Errs = map[string]*Perr{
		"connect": {
			100,
			Conn,
			"client connected",
			nil},
		"dispatch": {
			101,
			Debug,
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
			Debug,
			"Quit called: closing listener socket",
			nil},
		"success": {
			200,
			Debug,
			"reply sent",
			nil},
		"badreq": {
			400,
			Debug,
			"bad command",
			[]byte("PERRPERR400")},
		"nilreq": {
			401,
			Debug,
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

	// Errmap lets you go the other way, from a numeric status to
	// the name of a Perr
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

	// Loglvl maps string logging levels (from configurations) to
	// their int equivalents (actually used in code)
	Loglvl = map[string]int{
		"debug": Debug,
		"conn":  Conn,
		"error": Error,
		"fatal": Fatal}
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
