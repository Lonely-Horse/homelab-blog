package server

import "net/http"

func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handlerHealthz)
	mux.Handle("GET /admin", s.requireAuthPage(http.HandlerFunc(s.handlerAdmin)))
	mux.HandleFunc("GET /admin/login", s.handlerLoginPage)
	mux.HandleFunc("POST /admin/login", s.handlerLoginForm)
	mux.HandleFunc("POST /admin/logout", s.handlerLogout)

	mux.HandleFunc("POST /api/admin/login", s.handlerLoginAPI)

	return mux
}
