package main

import (
	"flag"
	"fmt"
	"homelab-blog/internal/config"
	"homelab-blog/internal/db"
	"homelab-blog/internal/markdown"
	"homelab-blog/internal/posts"
	"log"
)

func run(dryRun bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	database, err := db.Open(cfg)
	if err != nil {
		return err
	}
	defer database.Close()

	store := posts.NewStore(cfg, database)

	all, err := store.ListAll()
	if err != nil {
		return err
	}

	for _, p := range all {
		if p.RenderVersion >= cfg.ParserVersion {
			continue
		}

		body, headings := markdown.Render(p.ContentMD)
		toc := markdown.RenderTOC(headings)

		if dryRun {
			fmt.Printf("[INFO] The %s's length is %d\n", p.Slug, len(toc))
			continue
		}

		query := "UPDATE posts SET content_html = ?,toc_html = ?, render_version = ? WHERE id = ?"
		res, err := database.Exec(query, body, toc, cfg.ParserVersion, p.ID)
		if err != nil {
			return fmt.Errorf("post %s: %w", p.Slug, err)
		}

		affected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("post %s: %w", p.Slug, err)
		}
		if affected == 0 {
			return fmt.Errorf("post %d (%s): update affected 0 rows", p.ID, p.Slug)
		}
	}

	return nil
}

func main() {
	dryRun := flag.Bool("dry-run", false, "只统计不写库")
	flag.Parse()

	err := run(*dryRun)
	if err != nil {
		log.Fatalf("[ERROR] backfill failed: %v", err)
	}
}
