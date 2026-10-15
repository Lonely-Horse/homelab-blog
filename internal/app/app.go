package app

import (
	"homelab-blog/internal/auth"
	"homelab-blog/internal/config"
	"homelab-blog/internal/db"
	"homelab-blog/internal/posts"
	"homelab-blog/internal/profile"
	"homelab-blog/internal/projects"
	"homelab-blog/internal/server"
	"log"
	"os"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log.Printf("https is %v now", cfg.SessionCookieSecure)

	err = os.MkdirAll(cfg.UploadsDir, 0o755)
	if err != nil {
		return err
	}

	database, err := db.Open(cfg)
	if err != nil {
		log.Printf("The Open database: %s", err)
		return err
	}
	defer database.Close()

	authSvc := auth.NewService(cfg, database)
	postStore := posts.NewStore(cfg, database)
	profile := profile.NewStore(database)
	projects := projects.NewStore(database)
	srv, err := server.New(cfg, authSvc, postStore, profile, projects)
	if err != nil {
		log.Printf("The New server: %s", err)
		return err
	}
	return srv.Start()
}
