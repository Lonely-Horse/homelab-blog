package server

import (
	"bytes"
	"context"
	"errors"
	"homelab-blog/internal/auth"
	"homelab-blog/internal/config"
	"log"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"
)

// 登陆次数限流
type loginBucket struct {
	count       int
	windowStart time.Time
}

type loginLimiter struct {
	mu      sync.Mutex
	cfg     config.Config
	buckets map[string]*loginBucket
}

func newLoginLimiter(cfg config.Config) *loginLimiter {
	return &loginLimiter{
		cfg:     cfg,
		buckets: make(map[string]*loginBucket),
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return host
	}

	if addr.IsLoopback() {
		x := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		addr, err := netip.ParseAddr(x)
		if err != nil {
			return host
		}
		return addr.String()
	}

	return host
}

// 查询器
func (l *loginLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[ip]
	if !ok {
		return true
	}

	now := time.Now().UTC()
	expired := now.Sub(b.windowStart) >= l.cfg.LoginRateWindow // 只留一种写法

	if b.count >= l.cfg.LoginRatePerMin && !expired {
		return false
	}

	if expired {
		delete(l.buckets, ip)
	}

	return true
}

// 计数器
func (l *loginLimiter) recordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().UTC()

	b, ok := l.buckets[ip]
	if !ok {
		l.buckets[ip] = &loginBucket{count: 1, windowStart: now}
		return
	}

	expired := now.Sub(b.windowStart) >= l.cfg.LoginRateWindow // 只留一种写法

	if expired {
		l.buckets[ip] = &loginBucket{count: 1, windowStart: now} // 开一个新窗口
		return
	}
	b.count++
}

// 删除器
func (l *loginLimiter) reset(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.buckets, ip)
}

// 执行过期清理
func (l *loginLimiter) sweepExpired() {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now().UTC()

	for ip, bucket := range l.buckets {
		if now.Sub(bucket.windowStart) >= l.cfg.LoginRateWindow {
			delete(l.buckets, ip)
		}
	}

}

// 拉起进程，开始静候扫描
func (l *loginLimiter) startSweeper(ctx context.Context) {
	ticker := time.NewTicker(l.cfg.LoginRateWindow * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.sweepExpired()
		}
	}

}

var errTooManyRequests = errors.New("too_many_requests")

func (s *Server) tryLogin(r *http.Request, username, password string) (token string, err error) {
	ip := clientIP(r)

	if !s.limiter.allow(ip) {
		return "", errTooManyRequests
	}

	token, err = s.auth.Login(username, password)
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		s.limiter.recordFailure(ip)
		return "", err
	case err != nil:
		return "", err
	}

	s.limiter.reset(ip)

	return token, nil
}

type loginPageData struct {
	Error string
}

// 渲染登陆页面
func (s *Server) renderLogin(w http.ResponseWriter, status int, data loginPageData) {
	var buf bytes.Buffer

	err := s.tpl.ExecuteTemplate(&buf, "login.html", data)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	w.Write(buf.Bytes())

}

// 登陆逻辑实现
func (s *Server) handlerLoginAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"The URL is illegal"}`))
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"The username or password is empty"}`))
		return
	}

	token, err := s.tryLogin(r, username, password)
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_credentials"}`))
		return
	case errors.Is(err, errTooManyRequests):
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"too_many_requests"}`))
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal_error"}`))
		return
	}

	cookie := &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    token,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   s.cfg.SessionCookieSecure,
	}

	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handlerLoginForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		s.renderLogin(w, http.StatusBadRequest, loginPageData{Error: "The URL is illegal"})
		return
	}

	//获取并确定r中是否存在非空的字段
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	if username == "" || password == "" {
		s.renderLogin(w, http.StatusBadRequest, loginPageData{Error: "The username or password is empty"})
		return
	}

	token, err := s.tryLogin(r, username, password)
	switch {
	case errors.Is(err, errTooManyRequests):
		s.renderLogin(w, http.StatusTooManyRequests, loginPageData{Error: "尝试过于频繁，请稍候尝试"})
		return
	case errors.Is(err, auth.ErrInvalidCredentials):
		s.renderLogin(w, http.StatusUnauthorized, loginPageData{Error: "用户名或者密码错误"})
		return
	case err != nil:
		s.renderLogin(w, http.StatusInternalServerError, loginPageData{Error: "登陆失败"})
		return
	}

	cookie := &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    token,
		MaxAge:   int(s.cfg.SessionTTL.Seconds()),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   s.cfg.SessionCookieSecure,
	}

	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *Server) handlerLoginPage(w http.ResponseWriter, r *http.Request) {
	s.renderLogin(w, http.StatusOK, loginPageData{})
}

// 登出逻辑实现，本质就是删服务端和客户端session
func (s *Server) handlerLogout(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(s.cfg.SessionCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	err = s.auth.DeleteSession(token.Value)
	if err != nil {
		log.Printf("The delete session: %v", err)
		http.Error(w, "删除数据失败，请再次尝试", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.cfg.SessionCookieSecure,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
