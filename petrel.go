// Copyright (c) 2014-2025 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

package petrel

import (
	"os"
	"os/signal"
	"syscall"
)

// Status is a Petrel operational status. The data which is used to
// generate internal and external informational and error messages are
// stored as Statuses.
type Status struct {
	Lvl int
	Txt string
}

const (
	// Message levels control which messages will be sent to
	// h.Msgr, and the severity of Statuses
	Debug = iota
	Info
	Warn
	Error
	Fatal
)

var (
	// Proto is the version of the wire protocol implemented by
	// this library
	Proto = []byte{0}

	// Sigchan is the channel over which we listen for SIGs
	Sigchan chan os.Signal

	Loglvl = map[string]int{
		"debug": 0,
		"info" : 1,
		"warn" : 2,
		"error": 3,
		"fatal": 4,
	}
)

// Stats is the map of Status instances. It is used by Msg handling
// code throughout the Petrel packages, as a basis for the status of
// network responses, and to construct errors.
var Stats = map[uint16]*Status{
	100: {
		Info,
		"client connected",
	},
	101: {
		Debug,
		"dispatching",
	},
	198: {
		Info,
		"client disconnected",
	},
	199: {
		Debug,
		"Quit called: closing listener socket",
	},
	200: {
		Debug,
		"reply sent",
	},
	400: {
		Warn,
		"unknown command",
	},
	401: {
		Warn,
		"null command",
	},
	402: {
		Error,
		"payload length limit exceeded",
	},
	497: {
		Fatal,
		"protocol mismatch",
	},
	498: {
		Error,
		"network read error",
	},
	499: {
		Error,
		"network write error",
	},
	500: {
		Error,
		"request failed",
	},
	501: {
		Error,
		"internal error",
	},
	502: {
		Error,
		"HMAC verification failed",
	},
	599: {
		Fatal,
		"read from listener socket failed",
	},
}

func init() {
	// we'll listen for SIGINT and SIGTERM so we can behave like a
	// proper service. (mostly; we're not writing out a pidfile.)
	// we need a channel to receive signals on.
	Sigchan = make(chan os.Signal, 1)
	// and we need to register that channel to listen for and
	// respond properly to 'kill' calls to our pid, as well as to
	// C-c if running in a terminal.
	signal.Notify(Sigchan, syscall.SIGINT, syscall.SIGTERM)
}
