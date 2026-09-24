package main

import (
	"bufio"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"homelab-blog/internal/auth"
	"homelab-blog/internal/config"
	"homelab-blog/internal/db"
	"io"
	"os"
	"strings"
	"time"
)

func readPassword(br *bufio.Reader) (string, error) {
	line, err := br.ReadString('\n')
	if err != io.EOF && err != nil {
		return "", err
	}
	if line == "" {
		return "", fmt.Errorf("[ERROR] The password is empty")
	}

	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return line, nil
}

func run(user string, reset bool) error {
	if user == "" {
		return errors.New("The user is empty")
	}

	br := bufio.NewReader(os.Stdin)
	pwd, err := readPassword(br)
	if err != nil {
		return fmt.Errorf("[ERROR] Read password failed: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("[ERROR] The config model load failed,detail: %w", err)
	}

	hash, salt, err := auth.Hash(pwd, cfg.PBKDF2Iterations, cfg.PBKDF2KeyLength, cfg.SaltLength)
	if err != nil {
		return fmt.Errorf("[ERROR] The hash model used failed,detail: %w", err)
	}

	ok := auth.Verify(pwd, salt, hash, cfg.PBKDF2Iterations, cfg.PBKDF2KeyLength)
	if !ok {
		return errors.New("[ERROR] The hash and verify have problem")
	}

	database, err := db.Open(cfg)
	if err != nil {
		return fmt.Errorf("[ERROR] The database open failed,detail: %w", err)
	}
	defer database.Close()

	var id int
	query := "SELECT id FROM admins WHERE username = ?"
	err = database.QueryRow(query, user).Scan(&id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = database.Exec("INSERT INTO admins (username,password_hash,salt,iterations,created_at) VALUES (?,?,?,?,?)", user, hash, salt, cfg.PBKDF2Iterations, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("[ERROR] The database insert failed,detail: %w", err)
		}

	case err != nil:
		return fmt.Errorf("[ERROR] The query admin: %w", err)

	default:
		if !reset {
			return fmt.Errorf("[ERROR] The admin %q already exists", user)
		}
		fmt.Printf("[INFO] The password is reset")
	}

	return nil

}

func main() {
	var user string
	var reset bool
	flag.StringVar(&user, "user", "admin", "管理员名称")
	flag.BoolVar(&reset, "reset", false, "已存在时覆盖密码，默认拒绝")
	flag.Parse()

}
