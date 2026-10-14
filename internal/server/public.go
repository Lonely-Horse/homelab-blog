package server

import (
	"bytes"
	"errors"
	"homelab-blog/internal/posts"
	"homelab-blog/internal/projects"
	"html/template"
	"net/http"
	"time"
)

type IndexPageData struct {
	List []*posts.Post
}

type PostsPageData struct {
	Title       string
	Summary     string
	ContentHTML template.HTML
	UpdateAt    time.Time
}

type ErrorPageData struct {
	Status  int
	Message string
}

type AboutPageData struct {
	Posts       int
	ContentHTML template.HTML
	UpdatedAt   time.Time
	IsAdmin     bool
}

type ProjectPageData struct {
	List []*projects.Project
}

func (s *Server) handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) render(w http.ResponseWriter, statuscode int, name string, data any) {
	var buf bytes.Buffer

	err := s.tpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		http.Error(w, "internal Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statuscode)
	w.Write(buf.Bytes())
}

func (s *Server) handlerIndex(w http.ResponseWriter, r *http.Request) {
	list, err := s.posts.ListPublished()
	if err != nil {
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "Internal Server Error"})
		return
	}

	s.render(w, http.StatusOK, "index.html", IndexPageData{List: list})
}

func (s *Server) handlerPost(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	post, err := s.posts.GetBySlug(slug)
	switch {
	case errors.Is(err, posts.ErrNotFound):
		s.render(w, http.StatusNotFound, "error.html", ErrorPageData{Status: http.StatusNotFound, Message: "Post Not Found"})
		return
	case err != nil:
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "Internal Server Error"})
		return
	}

	if post.Status != "published" {
		s.render(w, http.StatusNotFound, "error.html", ErrorPageData{Status: http.StatusNotFound, Message: "Post Not Found"})
		return
	}

	s.render(w, http.StatusOK, "post.html", PostsPageData{Title: post.Title, Summary: post.Summary, ContentHTML: template.HTML(post.ContentHTML), UpdateAt: post.UpdatedAt})

}

func (s *Server) handlerDashboardPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, http.StatusOK, "dashboard.html", nil)
}

func (s *Server) handlerProjectsPage(w http.ResponseWriter, r *http.Request) {
	list, err := s.projects.ListPublished()
	if err != nil {
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
		return
	}

	s.render(w, http.StatusOK, "projects.html", ProjectPageData{List: list})
}
