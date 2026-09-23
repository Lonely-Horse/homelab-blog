package app

import (
	"homelab-blog/internal/config"
	"homelab-blog/internal/db"
	"homelab-blog/internal/markdown"
	"log"
	"net/http"
	"os"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.ReadFile("test.md")
		if err != nil {
			log.Printf("The error is %s", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("The markdown isn't exist"))
			return
		}
		result := markdown.Render(string(file))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(result))
	})

	log.Printf("The server listen on the %s", cfg.ListenAddr)
	http.ListenAndServe(cfg.ListenAddr, nil)

	return nil
}
