package posts

import (
	"database/sql"
	"errors"
	"homelab-blog/internal/config"
	"homelab-blog/internal/db"
	"time"
)

type Store struct {
	cfg config.Config
	db  *sql.DB
}

func NewStore(cfg config.Config, database *sql.DB) *Store {
	return &Store{cfg: cfg, db: database}
}

func (s *Store) Create(p *Post) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := db.NowUTC()

	//select 1这样的写法是存在性检查
	var exists int
	query1 := "SELECT 1 FROM posts WHERE slug = ?"
	err = tx.QueryRow(query1, p.Slug).Scan(&exists)
	switch {
	case errors.Is(err, sql.ErrNoRows):

	case err != nil:
		return 0, err
	case exists == 1:
		return 0, ErrDuplicate
	}

	if p.Status == "published" {
		p.PublishedAt = &now
	} else {
		p.PublishedAt = nil
	}

	p.CreatedAt = now
	p.UpdatedAt = now
	p.RenderVersion = s.cfg.ParserVersion

	query2 := "INSERT INTO posts(slug,title,summary,content_md,content_html,status,created_at,updated_at,published_at,render_version) VALUES (?,?,?,?,?,?,?,?,?,?)"
	res, err := tx.Exec(query2, p.Slug, p.Title, p.Summary, p.ContentMD, p.ContentHTML, p.Status, p.CreatedAt, p.UpdatedAt, p.PublishedAt, p.RenderVersion)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Store) Update(p *Post) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	query1 := "SELECT 1 FROM posts WHERE slug = ? AND id <> ?"
	err = tx.QueryRow(query1, p.Slug, p.ID).Scan(&exists)
	switch {
	case errors.Is(err, sql.ErrNoRows):

	case err != nil:
		return err
	case exists == 1:
		return ErrDuplicate
	}

	var oldStatus string
	var oldPublishedAt *time.Time
	query2 := "SELECT status,published_at FROM posts WHERE id = ?"
	err = tx.QueryRow(query2, p.ID).Scan(&oldStatus, &oldPublishedAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return ErrNotFound
	case err != nil:
		return err
	}

	now := db.NowUTC()
	p.RenderVersion = s.cfg.ParserVersion
	p.UpdatedAt = now
	switch {
	case p.Status == "published" && oldStatus != "published":
		p.PublishedAt = &now
	case p.Status != "published":
		p.PublishedAt = nil
	default:
		p.PublishedAt = oldPublishedAt
	}

	query3 := "UPDATE posts SET slug = ?,title = ?,summary = ?,content_md = ?,content_html = ?,status = ?,updated_at = ?,published_at = ?,render_version = ? WHERE id = ?"
	res, err := tx.Exec(query3, p.Slug, p.Title, p.Summary, p.ContentMD, p.ContentHTML, p.Status, p.UpdatedAt, p.PublishedAt, p.RenderVersion, p.ID)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(id int64) error {
	query := "DELETE FROM posts WHERE id = ?"
	res, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	ID, err := res.RowsAffected()
	switch {
	case err != nil:
		return err
	case ID == 0:
		return ErrNotFound
	}

	return nil
}

const selectPostCols = `id,slug,title,summary,content_md,content_html,status,created_at,updated_at,published_at,render_version`

func (s *Store) GetBySlug(slug string) (*Post, error) {
	post := &Post{}
	query := "SELECT " + selectPostCols + " FROM posts WHERE slug = ?"
	err := s.db.QueryRow(query, slug).Scan(&post.ID, &post.Slug, &post.Title, &post.Summary, &post.ContentMD, &post.ContentHTML, &post.Status, &post.CreatedAt, &post.UpdatedAt, &post.PublishedAt, &post.RenderVersion)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err != nil:
		return nil, err
	}

	return post, nil
}

func (s *Store) GetById(id int64) (*Post, error) {
	post := &Post{}
	query := "SELECT " + selectPostCols + " FROM posts WHERE id = ?"
	err := s.db.QueryRow(query, id).Scan(&post.ID, &post.Slug, &post.Title, &post.Summary, &post.ContentMD, &post.ContentHTML, &post.Status, &post.CreatedAt, &post.UpdatedAt, &post.PublishedAt, &post.RenderVersion)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err != nil:
		return nil, err
	}

	return post, nil
}

func (s *Store) ListPublished() ([]*Post, error) {
	query := "SELECT " + selectPostCols + " FROM posts WHERE status = 'published' ORDER BY published_at DESC"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(&post.ID, &post.Slug, &post.Title, &post.Summary, &post.ContentMD, &post.ContentHTML, &post.Status, &post.CreatedAt, &post.UpdatedAt, &post.PublishedAt, &post.RenderVersion)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (s *Store) ListAll() ([]*Post, error) {
	query := "SELECT " + selectPostCols + " FROM posts ORDER BY updated_at DESC"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err := rows.Scan(&post.ID, &post.Slug, &post.Title, &post.Summary, &post.ContentMD, &post.ContentHTML, &post.Status, &post.CreatedAt, &post.UpdatedAt, &post.PublishedAt, &post.RenderVersion)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return posts, nil
}
