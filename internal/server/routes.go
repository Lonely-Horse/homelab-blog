package server

import "net/http"

func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handlerHealthz)
	mux.Handle("GET /admin", s.requireAuthPage(http.HandlerFunc(s.handlerAdmin)))
	mux.HandleFunc("GET /admin/login", s.handlerLoginPage)
	mux.HandleFunc("POST /admin/login", s.handlerLoginForm)
	mux.HandleFunc("POST /admin/logout", s.handlerLogout)
	mux.Handle("POST /admin/posts", s.requireAuthPage(http.HandlerFunc(s.handlerCreatePost)))
	mux.Handle("GET /admin/posts/new", s.requireAuthPage(http.HandlerFunc(s.handlerNewPostPage)))
	mux.Handle("GET /admin/posts", s.requireAuthPage(http.HandlerFunc(s.handlerAdminShowPost)))
	mux.Handle("GET /admin/posts/{id}/edit", s.requireAuthPage(http.HandlerFunc(s.handlerEditPostPage)))
	mux.Handle("POST /admin/posts/{id}", s.requireAuthPage(http.HandlerFunc(s.handlerUpdatePost)))
	mux.Handle("DELETE /admin/posts/{id}", s.requireAuthAPI(http.HandlerFunc(s.handlerDeletePost)))

	mux.HandleFunc("GET /{$}", s.handlerIndex)
	mux.HandleFunc("GET /post/{slug}", s.handlerPost)

	mux.HandleFunc("POST /api/admin/login", s.handlerLoginAPI)

	return mux
}
