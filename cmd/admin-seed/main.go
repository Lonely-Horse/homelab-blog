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

	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	if strings.TrimSpace(line) == "" {
		return "", errors.New("The password is empty")
	}

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

	if !auth.Verify(pwd, salt, hash, cfg.PBKDF2Iterations, cfg.PBKDF2KeyLength) {
		return errors.New("The password hash isn't verify!")
	}

	database, err := db.Open(cfg)
	if err != nil {
		return fmt.Errorf("[ERROR] The database open failed,detail: %w", err)
	}
	defer database.Close()

	var id int64
	query := "SELECT id FROM admins WHERE username = ?"
	err = database.QueryRow(query, user).Scan(&id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = database.Exec("INSERT INTO admins (username,password_hash,salt,iterations,created_at) VALUES (?,?,?,?,?)", user, hash, salt, cfg.PBKDF2Iterations, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("[ERROR] The database insert failed,detail: %w", err)
		}
		fmt.Printf("[INFO] Admin %q created successfully\n", user)

	case err != nil:
		return fmt.Errorf("[ERROR] The query admin: %w", err)

	default:
		if !reset {
			return fmt.Errorf("[ERROR] The admin %q already exists", user)
		}

		tx, err := database.Begin()
		if err != nil {
			return fmt.Errorf("[ERROR] The begin: %w", err)
		}
		defer tx.Rollback()

		query2 := "UPDATE admins SET password_hash = ?, salt = ?, iterations = ? WHERE username = ?"
		_, err = tx.Exec(query2, hash, salt, cfg.PBKDF2Iterations, user)
		if err != nil {
			return fmt.Errorf("[ERROR] The update admin: %w", err)
		}

		query3 := "DELETE FROM sessions WHERE admin_id = ?"
		result, err := tx.Exec(query3, id)
		if err != nil {
			return fmt.Errorf("[ERROR] Other session delete: %w", err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("[ERROR] The commit: %w", err)
		}

		n, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("[ERROR] The delete session: %w", err)
		}

		fmt.Printf("[INFO] Admin %q update successfully\n", user)
		fmt.Printf("[INFO] The %d sessions delete successfully\n", n)
	}

	return nil

}

func main() {
	var user string
	var reset bool
	flag.StringVar(&user, "user", "admin", "管理员名称")
	flag.BoolVar(&reset, "reset", false, "已存在时覆盖密码，默认拒绝")
	flag.Parse()

	err := run(user, reset)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
