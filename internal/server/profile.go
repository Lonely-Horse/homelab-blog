package server

import (
	"homelab-blog/internal/markdown"
	"homelab-blog/internal/profile"
	"html/template"
	"net/http"
	"time"
)

func (s *Server) handlerAboutPage(w http.ResponseWriter, r *http.Request) {
	profile, err := s.profile.Get()
	if err != nil {
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Get failed"})
		return
	}

	content_html := markdown.Render(profile.ContentMD)

	// 只有已登录才在 /about 暴露管理入口（红线：入口不得泄露给匿名用户）
	isAdmin := s.checkAuth(r) != nil

	s.render(w, http.StatusOK, "about.html", AboutPageData{Posts: profile.Posts, ContentHTML: template.HTML(content_html), UpdatedAt: profile.UpdatedAt, IsAdmin: isAdmin})

}

func (s *Server) handlerAboutEditPage(w http.ResponseWriter, r *http.Request) {
	p, err := s.profile.Get()
	if err != nil {
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The Get failed"})
		return
	}

	s.render(w, http.StatusOK, "admin_about_edit.html", profile.Profile{ContentMD: p.ContentMD})

}

func (s *Server) handlerAboutSavePage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.render(w, http.StatusBadRequest, "error.html", ErrorPageData{Status: http.StatusBadRequest, Message: "The URL illegal"})
		return
	}

	content_md := r.PostForm.Get("content_md")

	err = s.profile.Save(&profile.Profile{ContentMD: content_md, UpdatedAt: time.Now().UTC()})
	if err != nil {
		s.render(w, http.StatusInternalServerError, "error.html", ErrorPageData{Status: http.StatusInternalServerError, Message: "The save failed"})
		return
	}

	http.Redirect(w, r, "/about", http.StatusSeeOther)

}
