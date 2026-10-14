package projects

import (
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("project not found")
	ErrDuplicate = errors.New("project slug already exists")
)

type Project struct {
	ID         int64
	Slug       string
	Title      string
	Summary    string
	GithubURL  string
	DemoURL    string
	TechStack  string
	CoverImage string
	Status     string
	Featured   bool
	SortOrder  int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
