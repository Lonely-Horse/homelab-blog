package server

import (
	"context"
	"homelab-blog/internal/auth"
	"homelab-blog/internal/config"
	"homelab-blog/internal/posts"
	"homelab-blog/internal/profile"
	"homelab-blog/internal/projects"
	"homelab-blog/web"
	"html/template"
	"net/http"
)

type server struct {
	cfg       config.Config
	authSvc   *auth.Service
	postStore *posts.Store
	profile   *profile.Store
	projects  *projects.Store
}

func New(server server) (*Server, error) {
	tpl, err := template.ParseFS(web.FS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{
		cfg:      server.cfg,
		auth:     server.authSvc,
		posts:    server.postStore,
		profile:  server.profile,
		projects: server.projects,
		limiter:  newLoginLimiter(server.cfg),
		tpl:      tpl,
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
