package server

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"homelab-blog/internal/markdown"
	"homelab-blog/internal/posts"
	"homelab-blog/internal/projects"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func parseProjectForm(r *http.Request) (*projects.Project, error) {
	title := strings.TrimSpace(r.PostForm.Get("title"))
	slug := strings.TrimSpace(r.PostForm.Get("slug"))
	summary := strings.TrimSpace(r.PostForm.Get("summary"))
	githubURL := strings.TrimSpace(r.PostForm.Get("github_url"))
	demoURL := strings.TrimSpace(r.PostForm.Get("demo_url"))
	coverImage := strings.TrimSpace(r.PostForm.Get("cover_image"))
	status := strings.TrimSpace(r.PostForm.Get("status"))
	techStack := strings.TrimSpace(r.PostForm.Get("tech_stack"))

	featured := r.PostForm.Has("featured")

	sortOrder, err := strconv.Atoi(strings.TrimSpace(r.PostForm.Get("sort_order")))
	if err != nil {
		sortOrder = 0
	}

	p := &projects.Project{
		Slug:       slug,
		Title:      title,
		Summary:    summary,
		GithubURL:  githubURL,
		DemoURL:    demoURL,
		CoverImage: coverImage,
		Status:     status,
		TechStack:  techStack,
		SortOrder:  sortOrder,
		Featured:   featured,
	}

	switch {
	case title == "":
		return p, errors.New("The title is empty")
	case slug == "":
		return p, errors.New("The slug is empty")
	case len(title) > 250:
		return p, errors.New("The title too long")
	case summary == "":
		return p, errors.New("The summary is empty")
	case !strings.HasPrefix(githubURL, "https://github.com/"):
		return p, errors.New("The github url is bad")
	case demoURL != "" && !strings.HasPrefix(demoURL, "https://"):
		return p, errors.New("The URL is illegal")
	case status != "active" && status != "archived":
		return p, errors.New("The status is illegal")
	}

	return p, nil
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
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxUploadBytes)
	err := r.ParseMultipartForm(s.cfg.UploadMemoryBytes)
	var maxErr *http.MaxBytesError
	switch {
	case errors.As(err, &maxErr):
		s.render(w, http.StatusRequestEntityTooLarge, "error.html", ErrorPageData{Status: http.StatusRequestEntityTooLarge, Message: "The body too large"})
		return
	case err != nil:
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The request is bad"})
		return
	}

	p, err := parseProjectForm(r)
	if err != nil {
		s.render(w, http.StatusBadRequest, "admin_project_edit.html", AdminProjectEditPageData{IsNew: true, FormAction: "/admin/projects", Error: err.Error(), Project: p})
		return
	}

	f, _, err := r.FormFile("archive")
	switch {
	case errors.Is(err, http.ErrMissingFile):

	case err != nil:
		s.render(w, http.StatusBadRequest, "admin_project_edit.html", AdminProjectEditPageData{IsNew: true, FormAction: "/admin/projects", Error: err.Error(), Project: p})
		return
	default:
		defer f.Close()
		head := make([]byte, 4)
		_, err = io.ReadFull(f, head)
		if err != nil {
			s.render(w, http.StatusBadRequest, "admin_project_edit.html", AdminProjectEditPageData{IsNew: true, FormAction: "/admin/projects", Error: "文件读取问题", Project: p})
			return
		}
		if !bytes.Equal(head, []byte{'P', 'K', 3, 4}) {
			s.render(w, http.StatusBadRequest, "admin_project_edit.html", AdminProjectEditPageData{IsNew: true, FormAction: "/admin/projects", Error: "文件读取问题", Project: p})
			return
		}

		buf := make([]byte, 16)
		_, err = io.ReadFull(rand.Reader, buf)
		if err != nil {
			s.render(w, http.StatusBadRequest, "admin_project_edit.html", AdminProjectEditPageData{IsNew: true, FormAction: "/admin/projects", Error: err.Error(), Project: p})
			return
		}

		name := hex.EncodeToString(buf) + ".zip"

		dest, err := os.Create(filepath.Join(s.cfg.UploadsDir, name))
		if err != nil {
			s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
			return
		}
		defer dest.Close()

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
			err := os.Remove(filepath.Join(s.cfg.UploadsDir, name))
			if err != nil {
				log.Printf("The %s remove failed,detail: %s", name, err)
			}
			return
		}

		_, err = io.Copy(dest, f)
		if err != nil {
			s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
			err := os.Remove(filepath.Join(s.cfg.UploadsDir, name))
			if err != nil {
				log.Printf("The %s remove failed,detail: %s", name, err)
			}
			return
		}
		p.ArchivePath = name
	}

	_, err = s.projects.Create(p)
	switch {
	case errors.Is(err, projects.ErrDuplicate):
		s.render(w, http.StatusConflict, "admin_project_edit.html", AdminProjectEditPageData{IsNew: true, FormAction: "/admin/projects", Error: "slug 已经存在，请换一个", Project: p})
		if p.ArchivePath != "" {
			err := os.Remove(filepath.Join(s.cfg.UploadsDir, p.ArchivePath))
			if err != nil {
				log.Printf("The %s remove failed,detail: %s", p.ArchivePath, err)
			}
		}
		return
	case err != nil:
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
		if p.ArchivePath != "" {
			err := os.Remove(filepath.Join(s.cfg.UploadsDir, p.ArchivePath))
			if err != nil {
				log.Printf("The %s remove failed,detail: %s", p.ArchivePath, err)
			}
		}
		return
	}

	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)

}

func (s *Server) handlerEditProjectPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The URL is illegal"})
		return
	}

	p, err := s.projects.GetById(id)
	switch {
	case errors.Is(err, projects.ErrNotFound):
		s.render(w, http.StatusNotFound, "error.html", ErrorPageData{Status: http.StatusNotFound, Message: "The projects isn't found"})
		return
	case err != nil:
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
		return
	}
	s.render(w, http.StatusOK, "admin_project_edit.html", AdminProjectEditPageData{IsNew: false, FormAction: fmt.Sprintf("/admin/projects/%d", id), Project: p})

}

func (s *Server) handlerUpdateProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The quest a bad"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxUploadBytes)
	err = r.ParseMultipartForm(s.cfg.UploadMemoryBytes)
	var maxErr *http.MaxBytesError
	switch {
	case errors.As(err, &maxErr):
		s.render(w, http.StatusRequestEntityTooLarge, "error.html", ErrorPageData{Status: http.StatusRequestEntityTooLarge, Message: "The body too large"})
		return
	case err != nil:
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The request is bad"})
		return
	}

	old, err := s.projects.GetById(id)
	switch {
	case errors.Is(err, projects.ErrNotFound):
		s.render(w, http.StatusNotFound, "error.html", ErrorPageData{Status: http.StatusNotFound, Message: "The projects isn't found"})
		return
	case err != nil:
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
		return
	}

	p, err := parseProjectForm(r)
	if err != nil {
		s.render(w, http.StatusBadRequest, "admin_project_edit.html", AdminProjectEditPageData{IsNew: false, FormAction: fmt.Sprintf("/admin/projects/%d", id), Error: err.Error(), Project: p})
		return
	}

	p.ID = id
	p.ArchivePath = old.ArchivePath
	saved := ""
	f, _, err := r.FormFile("archive")
	switch {
	case errors.Is(err, http.ErrMissingFile):

	case err != nil:
		s.render(w, http.StatusBadRequest, "admin_project_edit.html", AdminProjectEditPageData{IsNew: false, FormAction: fmt.Sprintf("/admin/projects/%d", id), Project: p})
		return
	default:
		defer f.Close()
		head := make([]byte, 4)
		_, err = io.ReadFull(f, head)
		if err != nil {
			s.render(w, http.StatusBadRequest, "admin_project_edit.html", AdminProjectEditPageData{IsNew: false, FormAction: fmt.Sprintf("/admin/projects/%d", id), Project: p})
			return
		}
		if !bytes.Equal(head, []byte{'P', 'K', 3, 4}) {
			s.render(w, http.StatusBadRequest, "admin_project_edit.html", AdminProjectEditPageData{IsNew: false, FormAction: fmt.Sprintf("/admin/projects/%d", id), Project: p})
			return
		}

		buf := make([]byte, 16)
		_, err = io.ReadFull(rand.Reader, buf)
		if err != nil {
			s.render(w, http.StatusBadRequest, "admin_project_edit.html", AdminProjectEditPageData{IsNew: false, FormAction: fmt.Sprintf("/admin/projects/%d", id), Project: p})
			return
		}

		name := hex.EncodeToString(buf) + ".zip"

		dest, err := os.Create(filepath.Join(s.cfg.UploadsDir, name))
		if err != nil {
			s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
			return
		}
		defer dest.Close()

		_, err = f.Seek(0, io.SeekStart)
		if err != nil {
			s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
			err := os.Remove(filepath.Join(s.cfg.UploadsDir, name))
			if err != nil {
				log.Printf("The %s remove failed,detail: %s", name, err)
			}
			return
		}

		_, err = io.Copy(dest, f)
		if err != nil {
			s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
			err := os.Remove(filepath.Join(s.cfg.UploadsDir, name))
			if err != nil {
				log.Printf("The %s remove failed,detail: %s", name, err)
			}
			return
		}
		saved = name
		p.ArchivePath = name
	}

	err = s.projects.Update(p)
	switch {
	case errors.Is(err, projects.ErrDuplicate):
		s.render(w, http.StatusConflict, "admin_project_edit.html", AdminProjectEditPageData{IsNew: false, FormAction: fmt.Sprintf("/admin/projects/%d", id), Error: "slug冲突，请换一个slug", Project: p})
		if saved != "" {
			err := os.Remove(filepath.Join(s.cfg.UploadsDir, p.ArchivePath))
			if err != nil {
				log.Printf("The %s remove failed,detail: %s", p.ArchivePath, err)
			}
		}
		return
	case errors.Is(err, projects.ErrNotFound):
		s.render(w, http.StatusNotFound, "error.html", ErrorPageData{Status: http.StatusNotFound, Message: "The projects isn't found"})
		if saved != "" {
			err := os.Remove(filepath.Join(s.cfg.UploadsDir, p.ArchivePath))
			if err != nil {
				log.Printf("The %s remove failed,detail: %s", p.ArchivePath, err)
			}
		}
		return
	case err != nil:
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The quest is bad"})
		if saved != "" {
			err := os.Remove(filepath.Join(s.cfg.UploadsDir, p.ArchivePath))
			if err != nil {
				log.Printf("The %s remove failed,detail: %s", p.ArchivePath, err)
			}
		}
		return
	}

	if saved != "" && old.ArchivePath != "" {
		err := os.Remove(filepath.Join(s.cfg.UploadsDir, old.ArchivePath))
		if err != nil {
			log.Printf("The %s remove failed,detail: %s", old.ArchivePath, err)
		}
	}

	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)

}

func (s *Server) handlerDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The resquest is bad"})
		return
	}

	err = s.projects.Delete(id)
	switch {
	case errors.Is(err, projects.ErrNotFound):
		s.render(w, http.StatusNotFound, "error.html", ErrorPageData{Status: http.StatusNotFound, Message: "The project isn't found"})
		return
	case err != nil:
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Internal Server Error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
