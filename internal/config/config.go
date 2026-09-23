package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	//服务配置
	ListenAddr        string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int

	//数据路径
	DataDir string
	DBPath  string

	// 会话
	SessionCookieName   string
	SessionTTL          time.Duration
	SessionTokenBytes   int
	SessionCookieSecure bool // 纯 HTTP=false；上了 HTTPS 必须 true

	// 限流
	LoginRatePerMin int
	LoginRateWindow time.Duration

	// 密码派生
	PBKDF2Iterations int
	PBKDF2KeyLength  int
	SaltLength       int
}

func Load() (Config, error) {
	cfg := Config{
		//server
		ListenAddr:        "127.0.0.1:5200",
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ShutdownTimeout:   10 * time.Second,
		MaxHeaderBytes:    1 << 20,

		//data
		DataDir: "./data",
		DBPath:  "./data/blog.db",

		//session
		SessionCookieName:   "session",
		SessionTTL:          7 * 24 * time.Hour,
		SessionTokenBytes:   32,
		SessionCookieSecure: false,
		//限流
		LoginRatePerMin: 5,
		LoginRateWindow: 60 * time.Second,

		//密码派生
		PBKDF2Iterations: 600000,
		PBKDF2KeyLength:  32,
		SaltLength:       16,
	}

	valueStr, ok := os.LookupEnv("SESSION_COOKIE_SECURE")
	if ok {
		value, err := strconv.ParseBool(valueStr)
		if err != nil {
			return cfg, fmt.Errorf("The strconv bool failed,detail: %s", err)
		}
		cfg.SessionCookieSecure = value
	}
	log.Printf("https is %v now", cfg.SessionCookieSecure)
	return cfg, nil
}
