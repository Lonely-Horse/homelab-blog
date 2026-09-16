package main

import (
	"log"

	"homelab-blog/internal/app"
)

func main() {
	err := app.Run()
	if err != nil {
		log.Fatalf("fatal: %v", err)
	}
}
