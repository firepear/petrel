//go:build testing

package server

// RemoveHandler allows the removal of Handlers from the server
// dispatch table. As its purpose is to allow testing of
// error-handling within the client, it is only compiled in and
// available when the `testing` build flag is provided.
func (s *Server) RemoveHandler(name string) bool {
	delete(s.d, name)
	_, ok := s.d[name]
	return !ok
}
