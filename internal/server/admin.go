package server

import (
	"errors"
	"fmt"
	"homelab-blog/internal/markdown"
	"homelab-blog/internal/posts"
	"log"
	"net/http"
	"strconv"
)

type AdminPostPageData struct {
	List []*posts.Post
}

type AdminEditPageData struct {
	IsNew      bool
	FormAction string
	Error      string
	Post       *posts.Post
}

func (s *Server) handlerAdmin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
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
	s.render(w, http.StatusOK, "admin_edit.html", AdminEditPageData{IsNew: true, FormAction: "/admin/posts", Error: "", Post: &posts.Post{}})
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
		s.render(w, http.StatusConflict, "admin_edit.html", AdminEditPageData{IsNew: true, FormAction: "/admin/posts", Error: "slug冲突，请换一个", Post: &posts.Post{Title: title, Slug: slug, Summary: summary, ContentMD: content_md}})
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

	s.render(w, http.StatusOK, "admin_edit.html", AdminEditPageData{
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
		s.render(w, http.StatusConflict, "admin_edit.html", AdminEditPageData{
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
