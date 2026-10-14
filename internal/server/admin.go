package server

import (
	"errors"
	"fmt"
	"homelab-blog/internal/markdown"
	"homelab-blog/internal/posts"
	"homelab-blog/internal/projects"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotFound  = errors.New("post not found")
	ErrDuplicate = errors.New("post slug already exists")
)

type AdminPostPageData struct {
	List []*posts.Post
}

type AdminPostEditPageData struct {
	IsNew      bool
	FormAction string
	Error      string
	Post       *posts.Post
}

func (s *Server) handlerAdmin(w http.ResponseWriter, r *http.Request) {
	s.render(w, http.StatusOK, "admin_home.html", nil)
}

func (s *Server) handlerAdminShowPost(w http.ResponseWriter, r *http.Request) {
	list, err := s.posts.ListAll()
	if err != nil {
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "Internal Server Error"})
		return
	}

	s.render(w, http.StatusOK, "admin_list.html", AdminPostPageData{List: list})
}

func (s *Server) handlerNewPostPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, http.StatusOK, "admin_edit.html", AdminPostEditPageData{IsNew: true, FormAction: "/admin/posts", Error: "", Post: &posts.Post{}})
}

func (s *Server) handlerCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The URL is illegal"})
		return
	}

	title := r.PostForm.Get("title")
	slug := r.PostForm.Get("slug")
	summary := r.PostForm.Get("summary")
	content_md := r.PostForm.Get("content_md")
	status := r.PostForm.Get("status")

	switch {
	case title == "":
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The title is empty"})
		return
	case slug == "":
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The slug is empty"})
		return
	case status != "draft" && status != "published":
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The status is illegal"})
		return
	}

	html := markdown.Render(content_md)
	_, err = s.posts.Create(&posts.Post{Title: title, Slug: slug, Summary: summary, ContentMD: content_md, ContentHTML: html, Status: status})
	switch {
	case errors.Is(err, posts.ErrDuplicate):
		s.render(w, http.StatusConflict, "admin_edit.html", AdminPostEditPageData{IsNew: true, FormAction: "/admin/posts", Error: "slug冲突，请换一个", Post: &posts.Post{Title: title, Slug: slug, Summary: summary, ContentMD: content_md}})
		return
	case err != nil:
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The post create failed"})
		return
	}

	log.Printf("The %s post created successfully", title)
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)

}

func (s *Server) handlerEditPostPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The id switch failed"})
		return
	}

	post, err := s.posts.GetById(id)
	switch {
	case errors.Is(err, posts.ErrNotFound):
		s.render(w, http.StatusNotFound, "error.html", ErrorPageData{Status: http.StatusNotFound, Message: "The Post Not Found"})
		return
	case err != nil:
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The Request Failed"})
		return
	}

	s.render(w, http.StatusOK, "admin_edit.html", AdminPostEditPageData{
		IsNew:      false,
		FormAction: fmt.Sprintf("/admin/posts/%d", id),
		Error:      "",
		Post:       post,
	})

}

func (s *Server) handlerUpdatePost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The URL is illegal"})
		return
	}

	err = r.ParseForm()
	if err != nil {
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The URL is illegal"})
		return
	}

	title := r.PostForm.Get("title")
	slug := r.PostForm.Get("slug")
	summary := r.PostForm.Get("summary")
	content_md := r.PostForm.Get("content_md")
	status := r.PostForm.Get("status")

	switch {
	case title == "":
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The title is empty"})
		return
	case slug == "":
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The slug is empty"})
		return
	case status != "draft" && status != "published":
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The status is illegal"})
		return
	}

	content_html := markdown.Render(content_md)

	err = s.posts.Update(&posts.Post{ID: id, Title: title, Slug: slug, Summary: summary, ContentMD: content_md, ContentHTML: content_html, Status: status})
	switch {
	case errors.Is(err, posts.ErrDuplicate):
		s.render(w, http.StatusConflict, "admin_edit.html", AdminPostEditPageData{
			IsNew:      false,
			FormAction: fmt.Sprintf("/admin/posts/%d", id),
			Error:      "博客冲突，请修改后再次尝试",
			Post: &posts.Post{
				Title:       title,
				Slug:        slug,
				Summary:     summary,
				ContentMD:   content_md,
				ContentHTML: content_html,
				Status:      status,
			},
		})
		return
	case errors.Is(err, posts.ErrNotFound):
		s.render(w, http.StatusNotFound, "error.html", ErrorPageData{Status: http.StatusNotFound, Message: "The Post Not Found"})
		return
	case err != nil:
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
		return
	}

	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

func (s *Server) handlerDeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The URL is illegal"})
		return
	}

	err = s.posts.Delete(id)
	switch {
	case errors.Is(err, posts.ErrNotFound):
		s.render(w, http.StatusNotFound, "error.html", ErrorPageData{Status: http.StatusNotFound, Message: "The Post Not Found"})
		return
	case err != nil:
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Request failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type AdminProjectPageData struct {
	List []*projects.Project
}

type AdminProjectEditPageData struct {
	IsNew      bool
	FormAction string
	Error      string
	Project    *projects.Project
}

var projectSlugRe = regexp.MustCompilePOSIX("^[a-z0-9-]+$")

func ParseProjectForm(r *http.Request) (*projects.Project, error) {
	title := strings.TrimSpace(r.PostForm.Get("title"))
	slug := strings.TrimSpace(r.PostForm.Get("slug"))
	summary := strings.TrimSpace(r.PostForm.Get("summary"))
	githubURl := strings.TrimSpace(r.PostForm.Get("github_url"))
	demoURl := strings.TrimSpace(r.PostForm.Get("demo_url"))
	coverImage := strings.TrimSpace(r.PostForm.Get("cover_image"))
	status := strings.TrimSpace(r.PostForm.Get("status"))
	techStack := strings.TrimSpace(r.PostForm.Get("tech_stack"))

	featured := r.PostForm.Has("featured")

	sortOrder, err := strconv.Atoi(strings.TrimSpace(r.PostForm.Get("sort_order")))
	switch {
	case err != nil:
		return nil, err
	case sortOrder == 0:
		return nil, errors.New("The sort_order not exists")
	}

	switch {
	case title == "":
		return nil, errors.New("The title is empty")
	case len(title) > 250:
		return nil, errors.New("The title too long")
	case summary == "":
		return nil, errors.New("The summary is empty")
	case !strings.HasPrefix(githubURl, "https://github.com/"):
		return nil, errors.New("The github url is bad")
	case demoURl != "" && !strings.HasPrefix(demoURl, "https://"):
		return nil, errors.New("The URL is illegal")
	case status != "active" && status != "archived":
		return nil, errors.New("The status is illegal")
	}

	return &projects.Project{
		Slug:       slug,
		Title:      title,
		Summary:    summary,
		GithubURL:  githubURl,
		DemoURL:    demoURl,
		CoverImage: coverImage,
		Status:     status,
		TechStack:  techStack,
		Featured:   featured,
	}, nil
}

func (s *Server) handlerAdminShowProject(w http.ResponseWriter, r *http.Request) {
	list, err := s.projects.ListAll()
	if err != nil {
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
		return
	}

	s.render(w, http.StatusOK, "admin_projects.html", AdminProjectPageData{List: list})

}

func (s *Server) handlerNewProjectPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, http.StatusOK, "admin_project_edit.html", AdminProjectEditPageData{IsNew: true, FormAction: "/admin/projects", Project: &projects.Project{Status: "active"}})
}

func (s *Server) handlerCreateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The quest bad"})
		return
	}

	p, err := ParseProjectForm(r)
	if err != nil {
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Inernal Server Error"})
		return
	}

	_, err = s.projects.Create(p)
	switch {
	case errors.Is(err, ErrDuplicate):
		s.render(w, http.StatusConflict, "admin_project_edit.html", AdminProjectEditPageData{Error: "slug 已经存在，请换一个", Project: p})
		return
	case err != nil:
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
		return
	}

	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)

}

func (s *Server) handlerEditProjectPage(w http.ResponseWriter, r *http.Request) {

}
