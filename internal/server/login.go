package server

import (
	"errors"
	"homelab-blog/internal/auth"
	"homelab-blog/internal/config"
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

// 登陆handler设计
func (s *Server) handlerLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Path is illegal"}`))
		return
	}

	//获取并确定r中是否存在非空的字段
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"username or password is empty"}`))
		return
	}

	ip := clientIP(r)
	if !s.limiter.allow(ip) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"The IP is too many request"}`))
		return
	}

	token, err := s.auth.Login(username, password)
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		s.limiter.recordFailure(ip)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid username or password"}`))
		return

	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Login failed"}`))
		return
	}

	s.limiter.reset(ip)

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

	http.Redirect(w, r, "/admin", http.StatusFound)
	w.Write([]byte(`{"status":"ok"}`))

}

func (s *Server) handlerLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := r.Cookie(s.cfg.SessionCookieName)
	switch {
	case errors.Is(err, http.ErrNoCookie):
		http.Redirect(w, r, "/admin/login", http.StatusFound)
		w.Write([]byte(`{"error":"cookie is empty"}`))
		return

	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"get cookie failed"}`))
		return
	}

	err = s.auth.DeleteSession(token.Value)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"delete session failed"}`))
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

	http.Redirect(w, r, "/admin/login", http.StatusFound)
	w.Write([]byte(`{"status":"ok"}`))

}
