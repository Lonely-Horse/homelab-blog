package server

import "net/http"

func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handlerHealthz)
	mux.Handle("GET /admin", s.requireAuthPage(http.HandlerFunc(s.handlerAdmin)))
	mux.HandleFunc("POST /admin/login", s.handlerLogin)
	mux.HandleFunc("POST /admin/logout", s.handlerLogout)
	return mux
}
