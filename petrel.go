// Copyright (c) 2014-2025 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

package petrel

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
		"in dispatch",
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
		"handler not found",
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
