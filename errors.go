package petrel

import "errors"

var (
	perrs = map[string]error{
		"reqlen": errors.New("request over limit; closing conn"),
	}
	perrb = map[string][]byte{
		"reqlen": []byte("PERRPERR402 Request over limit"),
		"default": []byte("PERRPERRAn error has occurred. Closing connection."),
	}
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
