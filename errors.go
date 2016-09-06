package petrel

import (
	"fmt"
)

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
	errcmderr = fmt.Errorf("dispatch cmd errored")
)
