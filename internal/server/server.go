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

type S struct {
	Cfg       config.Config
	AuthSvc   *auth.Service
	PostStore *posts.Store
	Profile   *profile.Store
	Projects  *projects.Store
}

func New(server S) (*Server, error) {
	tpl, err := template.ParseFS(web.FS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{
		cfg:      server.Cfg,
		auth:     server.AuthSvc,
		posts:    server.PostStore,
		profile:  server.Profile,
		projects: server.Projects,
		limiter:  newLoginLimiter(server.Cfg),
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
