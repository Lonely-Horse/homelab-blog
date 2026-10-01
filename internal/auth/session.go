package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"homelab-blog/internal/config"
)

type Service struct {
	cfg config.Config
	db  *sql.DB
}

func NewService(cfg config.Config, db *sql.DB) *Service {
	return &Service{cfg: cfg, db: db}
}

var ErrInvalidCredentials = errors.New("invalid username or password")

func generateToken(n int) (string, error) {
	token := make([]byte, n)
	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("[ERROR] The token rand: %w", err)
	}

	tokenstr := hex.EncodeToString(token)
	fmt.Printf("[INFO] The session token created successful")

	return tokenstr, nil
}

func (s *Service) CreateSession(adminID int64)
