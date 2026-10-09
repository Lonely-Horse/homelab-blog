package server

import (
	"context"
	"homelab-blog/internal/auth"
	"homelab-blog/internal/config"
	"homelab-blog/web"
	"html/template"
	"net/http"
)

func New(cfg config.Config, authSvc *auth.Service) (*Server, error) {
	tpl, err := template.ParseFS(web.FS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{
		cfg:     cfg,
		auth:    authSvc,
		limiter: newLoginLimiter(cfg),
		tpl:     tpl,
	}, nil
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.limiter.startSweeper(ctx)

	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
