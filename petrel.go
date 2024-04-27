// Copyright (c) 2014-2022 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

package petrel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	// Proto is the version of the wire protocol implemented by
	// this library
	Proto = uint8(0)

	// Message levels control which messages will be sent to
	// h.Msgr
	Debug = iota
	Conn
	Error
	Fatal
)

// Status is a Petrel operational status. The data which is used to
// generate internal and external informational and error messages are
// stored as Statuses.
type Status struct {
	Code int
	Lvl  int
	Txt  string
	Xmit []byte
}

// Stats is the map of Status instances. It is used by Msg handling
// code throughout the Petrel packages.
var Stats = map[string]*Status{
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
		"payload length limit exceeded; closing conn",
		[]byte("PERRPERR402")},
	"netreaderr": {
		498,
		Conn,
		"network read error",
		nil},
	"netwriteerr": {
		499,
		Conn,
		"network write error",
		nil},
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

// Loglvl maps string logging levels (from configurations) to
// their int equivalents (actually used in code)
var Loglvl = map[string]int{
	"debug": Debug,
	"conn":  Conn,
	"error": Error,
	"fatal": Fatal}

var (
	pverbuf = new(bytes.Buffer)
	Sigchan chan os.Signal
)

func init() {
	// pre-compute the binary encoding of Proto
	binary.Write(pverbuf, binary.LittleEndian, Proto)

	// we'll listen for SIGINT and SIGTERM so we can behave like a
	// proper service. (mostly; we're not writing out a pidfile.)
	// we need a channel to receive signals on.
	Sigchan = make(chan os.Signal, 1)

	// and we need to register that channel to listen for and
	// respond properly to 'kill' calls to our pid, as well as to
	// C-c if running in a terminal.
	signal.Notify(Sigchan, syscall.SIGINT, syscall.SIGTERM)
}

// Error implements the error interface for Status.
func (p Status) Error() string {
	return fmt.Sprintf("%s (%d)", p.Txt, p.Code)
}
