// genMsg creates messages and sends them to the Msgr channel.
func (s *Server) genMsg(conn, req uint32, stat uint16, err error) {
	// if this message's level is below the instance's level, don't
	// generate the message
	if p.Lvl < s.ml {
		return
	}
	s.Msgr <- &Msg{conn, req, stat, err}
}

// Msg is the format which Petrel uses to communicate informational
// messages and errors to its host program via the s.Msgr channel.
type Msg struct {
	// Conn is the connection ID that the Msg is coming from.
	Conn uint32
	// Req is the request number that resulted in the Msg.
	Req uint32
	// Code is the numeric status indicator.
	Code int16
	// Err is the error (if any) passed upward as part of the Msg.
	Err error
}

// Error implements the error interface for Msg, returning a nicely
// (if blandly) formatted string containing all information present.
func (m *Msg) Error() string {
	err := fmt.Sprintf("conn %d req %d status %d", m.Conn, m.Req, m.Code)
	if m.Txt != "" {
		err = err + fmt.Sprintf(" (%s)", m.Txt)
	}
	if m.Err != nil {
		err = err + fmt.Sprintf("; err: %s", m.Err)
	}
	return err
}
