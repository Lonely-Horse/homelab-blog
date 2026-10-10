package posts

import (
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("post not found")
	ErrDuplicate = errors.New("post slug already exists")
)

type Post struct {
	ID            int64
	Slug          string
	Title         string
	Summary       string
	ContentMD     string
	ContentHTML   string
	Status        string // "draft" | "published"
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PublishedAt   *time.Time // 可空：仅 published 非空
	RenderVersion int
}
