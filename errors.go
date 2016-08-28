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
		"netreaderr": &perr{
			196,
			Conn,
			"network read error",
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
		"nilreq": &perr{
			401,
			All,
			"nil request"
			[]byte("PERRPERR401received empty request") },
		"reqlen": &perr{
			402,
			All,
			"request over limit; closing conn",
			[]byte("PERRPERR402request over limit") },
		"internalerr": &perr{
			501,
			Error,
			"internal error",
			nil },
		"listenerfail": &perr{
			599,
			All,
			"read from listener socket failed",
			nil },
	}

	// these errors are for internal signalling; they do not propagate
	errbadcmd = fmt.Errorf("bad command")
	errcmderr = fmt.Errorf("dispatch cmd errored")
)

/*
    Code Text                                      Type
    ---- ----------------------------------------- -------------
     100 client connected                          Informational
     101 dispatching '%v'                                "
     196 network error                                   "
     197 ending session                                  "
     198 client disconnected                             "
     199 terminating listener socket                     "
     200 reply sent                                Success
     400 bad command '%v'                          Client error
     401 nil request                                     "
     402 request over limit                              "
     500 request failed                            Server Error
     501 internal error                                  "
     599 read from listener socket failed                "
*/
