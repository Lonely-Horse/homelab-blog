package projects

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound  = errors.New("project not found")
	ErrDuplicate = errors.New("project slug already exists")
)

type Project struct {
	ID          int64
	Slug        string
	Title       string
	Summary     string
	GithubURL   string
	DemoURL     string
	ArchivePath string
	TechStack   string
	CoverImage  string
	Status      string
	Featured    bool
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Tags 把存储的 tech_stack（形如 ",Go,Docker,"）解析为标签切片，供模板渲染。
// 铁律：先 Trim 两端逗号再 Split；空串返回 nil，避免渲染出空标签。
func (p *Project) Tags() []string {
	s := strings.Trim(p.TechStack, ",")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	tags := make([]string, 0, len(parts))
	for _, t := range parts {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
