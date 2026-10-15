package projects

import (
	"database/sql"
	"errors"
	"homelab-blog/internal/db"
)

type Store struct {
	db *sql.DB
}

func NewStore(database *sql.DB) *Store {
	return &Store{db: database}
}

func (s *Store) Create(p *Project) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var exists int
	query1 := "SELECT 1 FROM projects WHERE slug = ?"
	err = tx.QueryRow(query1, p.Slug).Scan(&exists)
	switch {
	case errors.Is(err, sql.ErrNoRows):

	case err != nil:
		return 0, err
	case exists == 1:
		return 0, ErrDuplicate
	}

	now := db.NowUTC()
	p.UpdatedAt = now
	p.CreatedAt = now

	query2 := "INSERT INTO projects(slug,title,summary,github_url,demo_url,archive_path,tech_stack,cover_image,status,featured,sort_order,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)"
	res, err := tx.Exec(query2, p.Slug, p.Title, p.Summary, p.GithubURL, p.DemoURL, p.ArchivePath, p.TechStack, p.CoverImage, p.Status, p.Featured, p.SortOrder, p.CreatedAt, p.UpdatedAt)
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

func (s *Store) Update(p *Project) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists int
	query1 := "SELECT 1 FROM projects WHERE slug = ? AND id <> ?"
	err = tx.QueryRow(query1, p.Slug, p.ID).Scan(&exists)
	switch {
	case errors.Is(err, sql.ErrNoRows):

	case err != nil:
		return err
	case exists == 1:
		return ErrDuplicate
	}

	now := db.NowUTC()
	p.UpdatedAt = now
	query2 := "UPDATE projects SET slug = ?,title = ?,summary = ?,github_url = ?,demo_url = ?,archive_path = ?,tech_stack = ?,cover_image = ?,status = ?,featured = ?,sort_order = ?,updated_at = ? WHERE id =?"
	res, err := tx.Exec(query2, p.Slug, p.Title, p.Summary, p.GithubURL, p.DemoURL, p.ArchivePath, p.TechStack, p.CoverImage, p.Status, p.Featured, p.SortOrder, p.UpdatedAt, p.ID)
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

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(id int64) error {
	query := "DELETE FROM projects WHERE id = ?"
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

const selectProjectsCols = `id,slug,title,summary,github_url,demo_url,archive_path,tech_stack,cover_image,status,featured,sort_order,created_at,updated_at`

func (s *Store) GetById(id int64) (*Project, error) {
	p := &Project{}
	query := "SELECT " + selectProjectsCols + " FROM projects WHERE id = ?"
	err := s.db.QueryRow(query, id).Scan(&p.ID, &p.Slug, &p.Title, &p.Summary, &p.GithubURL, &p.DemoURL, &p.ArchivePath, &p.TechStack, &p.CoverImage, &p.Status, &p.Featured, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err != nil:
		return nil, err
	}

	return p, nil
}

func (s *Store) ListAll() ([]*Project, error) {
	query := "SELECT " + selectProjectsCols + " FROM projects ORDER BY updated_at DESC"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*Project{}
	for rows.Next() {
		p := &Project{}
		err = rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Summary, &p.GithubURL, &p.DemoURL, &p.ArchivePath, &p.TechStack, &p.CoverImage, &p.Status, &p.Featured, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *Store) ListPublished() ([]*Project, error) {
	query := "SELECT " + selectProjectsCols + " FROM projects ORDER BY (status = 'archived') ASC,featured DESC,sort_order ASC,created_at DESC"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*Project{}
	for rows.Next() {
		p := &Project{}
		err = rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Summary, &p.GithubURL, &p.DemoURL, &p.ArchivePath, &p.TechStack, &p.CoverImage, &p.Status, &p.Featured, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return projects, nil
}
