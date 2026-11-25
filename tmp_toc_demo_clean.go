//go:build ignore

// 临时清理脚本：删除 TOC 演示文章（slug=toc-demo）。
// 用法：go run tmp_toc_demo_clean.go
package main

import (
	"fmt"
	"homelab-blog/internal/config"
	"homelab-blog/internal/db"
	"homelab-blog/internal/posts"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	database, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	store := posts.NewStore(cfg, database)
	post, err := store.GetBySlug("toc-demo")
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Delete(post.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted toc-demo, id =", post.ID)
}
