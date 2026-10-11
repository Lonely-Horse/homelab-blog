package server

import "net/http"

func (s *Server) handlerGetAbout(w http.ResponseWriter, r *http.Request) {
	_, err := s.profile.Get()
	if err != nil {
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Get failed"})
		return
	}

}
