package petrel

// Copyright (c) 2014-2016 Shawn Boyette <shawn@firepear.net>. All
// rights reserved.  Use of this source code is governed by a
// BSD-style license that can be found in the LICENSE file.

type perr struct {
	code int
	lvl  int
	txt  string
	xmit []byte
}

var (
	perrs = map[string]*perr{
		"connect": &perr{
			100,
			Conn,
			"client connected",
			nil },
		"dispatch": &perr{
			101,
			All,
			"dispatching",
			nil },
		"netreaderr": &perr{
			196,
			Conn,
			"network read error",
			nil },
		"netwriteerr": &perr{
			197,
			Conn,
			"network write error",
			nil },
		"disconnect": &perr{
			198,
			Conn,
			"client disconnected",
			nil },
		"quit": &perr{
			199,
			All,
			"Quit called: closing listener socket",
			nil },
		"success": &perr{
			200,
			All,
			"reply sent",
			nil },
		"badreq": &perr{
			400,
			All,
			"bad command",
			[]byte("PERRPERR400unknown command") },
		"nilreq": &perr{
			401,
			All,
			"nil request",
			[]byte("PERRPERR401received empty request") },
		"reqlen": &perr{
			402,
			All,
			"request over limit; closing conn",
			[]byte("PERRPERR402request over limit") },
		"reqerr": &perr{
			500,
			Error,
			"request failed",
			[]byte{"PERRPERR500request could not be completed"} },
		"internalerr": &perr{
			501,
			Error,
			"internal error",
			nil },
		"listenerfail": &perr{
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
		402: "reqlen",
		500: "reqerr",
		501: "internalerr",
		599: "listenerfail" }
)
