package app

import (
	"homelab-blog/internal/auth"
	"homelab-blog/internal/config"
	"homelab-blog/internal/db"
	"homelab-blog/internal/server"
	"log"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	database, err := db.Open(cfg)
	if err != nil {
		log.Printf("The error is %s", err)
		return err
	}
	defer database.Close()

	authSvc := auth.NewService(cfg, database)
	srv := server.New(cfg, authSvc)

	return srv.Start()
}
