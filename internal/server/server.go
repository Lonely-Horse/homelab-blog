package server

import (
	"homelab-blog/internal/auth"
	"homelab-blog/internal/config"
	"net/http"
)

func New(cfg config.Config, authSvc *auth.Service) *Server {
	return &Server{
		cfg:     cfg,
		auth:    authSvc,
		limiter: newLoginLimiter(cfg),
	}
}

func (s *Server) Start() error {
	server := &http.Server{
		Addr:              s.cfg.ListenAddr,
		Handler:           s.securityHeaders(s.routes()),
		ReadHeaderTimeout: s.cfg.ReadHeaderTimeout,
		ReadTimeout:       s.cfg.ReadTimeout,
		WriteTimeout:      s.cfg.WriteTimeout,
		IdleTimeout:       s.cfg.IdleTimeout,
	}

	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
