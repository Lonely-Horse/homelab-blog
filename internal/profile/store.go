package profile

import (
	"database/sql"
	"errors"
)

type Store struct {
	db *sql.DB
}

func NewStore(database *sql.DB) *Store {
	return &Store{db: database}
}

func (s *Store) Get() (*Profile, error) {
	profile := Profile{}
	query1 := "SELECT content_md,updated_at FROM site_profile WHERE id = 1"
	err := s.db.QueryRow(query1).Scan(&profile.ContentMD, &profile.UpdatedAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):

	case err != nil:
		return nil, err
	}

	query2 := "SELECT COUNT(*) FROM posts WHERE status = ?"
	err = s.db.QueryRow(query2, "published").Scan(&profile.Posts)
	if err != nil {
		return nil, err
	}

	query3 := "SELECT COUNT(*) FROM projects"
	err = s.db.QueryRow(query3).Scan(&profile.Projects)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (s *Store) Save(p *Profile) error {
	query := "INSERT INTO site_profile(id,content_md,updated_at) VALUES (1,?,?) ON CONFLICT(id) DO UPDATE SET content_md = excluded.content_md,updated_at = excluded.updated_at"
	_, err := s.db.Exec(query, p.ContentMD, p.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}
