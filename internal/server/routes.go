package server

import (
	"io/fs"
	"net/http"

	"homelab-blog/web"
)

func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handlerHealthz)

	// 静态资源：只暴露 embed 里的 static 子树，绝不能把整棵 FS（含 templates）挂出去
	staticFS, err := fs.Sub(web.FS, "static")
	if err != nil {
		// //go:embed static 已在编译期保证目录存在，走到这里说明 embed 指令被改坏
		panic("web: static 子树缺失: " + err.Error())
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	mux.HandleFunc("GET /{$}", s.handlerDashboardPage)
	mux.Handle("GET /admin", s.requireAuthPage(http.HandlerFunc(s.handlerAdmin)))
	mux.HandleFunc("GET /admin/login", s.handlerLoginPage)
	mux.HandleFunc("POST /admin/login", s.handlerLoginForm)
	mux.HandleFunc("POST /admin/logout", s.handlerLogout)
	mux.Handle("POST /admin/posts", s.requireAuthPage(http.HandlerFunc(s.handlerCreatePost)))
	mux.Handle("GET /admin/posts/new", s.requireAuthPage(http.HandlerFunc(s.handlerNewPostPage)))
	mux.Handle("GET /admin/posts", s.requireAuthPage(http.HandlerFunc(s.handlerAdminShowPost)))
	mux.Handle("GET /admin/posts/{id}/edit", s.requireAuthPage(http.HandlerFunc(s.handlerEditPostPage)))
	mux.Handle("POST /admin/posts/{id}", s.requireAuthPage(http.HandlerFunc(s.handlerUpdatePost)))
	mux.Handle("DELETE /admin/posts/{id}", s.requireAuthAPI(http.HandlerFunc(s.handlerDeletePost)))
	mux.Handle("GET /admin/projects", s.requireAuthPage(http.HandlerFunc(s.handlerAdminShowProject)))
	mux.Handle("GET /admin/projects/new", s.requireAuthPage(http.HandlerFunc(s.handlerNewProjectPage)))
	mux.Handle("POST /admin/projects", s.requireAuthPage(http.HandlerFunc(s.handlerCreateProject)))
	mux.Handle("GET /admin/projects/{id}/edit", s.requireAuthPage(http.HandlerFunc(s.handlerEditProjectPage)))
	mux.Handle("POST /admin/projects/{id}", s.requireAuthPage(http.HandlerFunc(s.handlerUpdateProject)))
	mux.Handle("DELETE /admin/projects/{id}", s.requireAuthAPI(http.HandlerFunc(s.handlerDeleteProject)))
	mux.Handle("POST /admin/preview", s.requireAuthAPI(http.HandlerFunc(s.handlerPreview)))

	mux.HandleFunc("GET /projects/{id}/download", s.handlerDownloadProject)
	mux.HandleFunc("GET /projects", s.handlerProjectsPage)
	mux.HandleFunc("GET /about", s.handlerAboutPage)
	mux.Handle("POST /about/edit", s.requireAuthPage(http.HandlerFunc(s.handlerAboutSavePage)))
	mux.Handle("GET /about/edit", s.requireAuthPage(http.HandlerFunc(s.handlerAboutEditPage)))
	mux.HandleFunc("GET /posts", s.handlerIndex)
	mux.HandleFunc("GET /post/{slug}", s.handlerPost)

	mux.HandleFunc("POST /api/admin/login", s.handlerLoginAPI)

	return mux
}
