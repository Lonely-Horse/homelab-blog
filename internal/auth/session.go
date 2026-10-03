package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"homelab-blog/internal/config"
	"log"
	"time"
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
		return "", fmt.Errorf("[ERROR] token rand: %w", err)
	}

	tokenstr := hex.EncodeToString(token)

	return tokenstr, nil
}

func (s *Service) CreateSession(adminID int64) (string, error) {
	token, err := generateToken(s.cfg.SessionTokenBytes)
	if err != nil {
		return "", fmt.Errorf("[ERROR] The generate token: %w", err)
	}

	now := time.Now().UTC()

	query := "INSERT INTO sessions (token,admin_id,expires_at,created_at) VALUES (?,?,?,?)"

	_, err = s.db.Exec(query, token, adminID, now.Add(s.cfg.SessionTTL), now)
	if err != nil {
		return "", fmt.Errorf("[ERROR] The session insert failed: %w", err)
	}

	return token, nil
}

func (s *Service) ValidSession(token string) (int64, bool) {
	var adminID int64
	var expiresAt time.Time
	query1 := "SELECT admin_id,expires_at FROM sessions WHERE token = ?"

	err := s.db.QueryRow(query1, token).Scan(&adminID, &expiresAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, false

	case err != nil:
		log.Printf("[ERROR] The select session: %v", err)
		return 0, false
	}

	if expiresAt.Before(time.Now().UTC()) {
		return 0, false
	}

	return adminID, true
}

func (s *Service) DeleteSession(token string) error {
	query := "DELETE FROM sessions WHERE token = ?"
	_, err := s.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("[ERROR] The delete session: %w", err)
	}

	return nil
}

func (s *Service) CleanExpired() error {
	query := "DELETE FROM sessions WHERE expires_at < ?"
	_, err := s.db.Exec(query, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("[ERROR] The cleanexpired: %w", err)
	}

	return nil
}

func (s *Service) Login(username, password string) (string, error) {
	var (
		adminID    int64
		storedHash []byte
		storedSalt []byte
		iterations int
	)

	query := "SELECT id,password_hash,salt,iterations FROM admins WHERE username = ?"
	err := s.db.QueryRow(query, username).Scan(&adminID, &storedHash, &storedSalt, &iterations)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, _, _ = Hash(password, s.cfg.PBKDF2Iterations, s.cfg.PBKDF2KeyLength, s.cfg.SaltLength)
		return "", ErrInvalidCredentials

	case err != nil:
		return "", fmt.Errorf("[ERROR] The query admin: %w", err)
	}

	if !Verify(password, storedSalt, storedHash, iterations, s.cfg.PBKDF2KeyLength) {
		return "", ErrInvalidCredentials
	}

	err = s.CleanExpired()
	if err != nil {
		log.Printf("The clean expired session: %v", err)
	}

	return s.CreateSession(adminID)
}
