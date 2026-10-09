package server

import (
	"context"
	"homelab-blog/internal/auth"
	"homelab-blog/internal/config"
	"html/template"
	"net/http"
	"strings"
)

type ctxKey int

const ctxKeyAdminID ctxKey = iota

type Server struct {
	cfg     config.Config
	auth    *auth.Service
	limiter *loginLimiter
	tpl     *template.Template
}

// 检查cookie中的值是否符合数据库中的合法值
func (s *Server) checkAuth(r *http.Request) *http.Request {
	cookie, err := r.Cookie(s.cfg.SessionCookieName)
	if err != nil {
		return nil
	}

	adminID, ok := s.auth.ValidSession(cookie.Value)
	if !ok {
		return nil
	}

	ctx := context.WithValue(r.Context(), ctxKeyAdminID, adminID)

	return r.WithContext(ctx)
}

// 将page和api更新为数据库中的合法token
func (s *Server) requireAuthPage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := s.checkAuth(r)
		if r2 == nil {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r2)
	})
}

func (s *Server) requireAuthAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := s.checkAuth(r)
		if r2 == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next.ServeHTTP(w, r2)
	})
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https: data:; style-src 'self'; script-src 'self'; form-action 'self'; base-uri 'self'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if strings.HasPrefix(r.URL.Path, "/admin") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}
