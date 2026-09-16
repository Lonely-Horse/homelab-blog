package app

import (
	"homelab-blog/internal/config"
	"homelab-blog/internal/db"
)

func Run() error {
	cfg := config.Load()

	database, err := db.Open(cfg)
	if err != nil {
		return err
	}
	defer database.Close()

	return nil
}
