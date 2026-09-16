package db

import (
	"database/sql"
	"fmt"
	"homelab-blog/internal/config"
	"log"
	"os"

	_ "embed"

	_ "github.com/glebarez/go-sqlite"
)

const schemaVersion = 1

//go:embed schema.sql
var schemaSQL string

func checkPragmas(db *sql.DB) error {
	var mode string

	err := db.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if err != nil {
		return err
	}
	log.Printf("The journal_mode: %s", mode)
	if mode != "wal" {
		return fmt.Errorf("The journal_mode=%q,want wal", mode)
	}

	var fk int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if err != nil {
		return err
	}
	log.Printf("The foreign_keys: %d", fk)
	if fk != 1 {
		return fmt.Errorf("The foreign_keys=%d,want 1", fk)
	}

	var sync int
	err = db.QueryRow("PRAGMA synchronous").Scan(&sync)
	if err != nil {
		return err
	}
	log.Printf("The synchronous=%d", sync)
	if sync != 1 {
		return fmt.Errorf("The synchronous=%d,want 1", sync)
	}

	return nil
}

func migrate(db *sql.DB) error {
	var current int
	err := db.QueryRow("PRAGMA user_version").Scan(&current)
	if err != nil {
		return err
	}

	log.Printf("schema migrate: user_version=%d, target=%d", current, schemaVersion)

	if current >= schemaVersion {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query("PRAGMA table_info(posts)")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasCol := false

	for rows.Next() {
		var (
			cid     int
			name    string
			typ     string
			notNull int
			dflt    sql.NullString
			pk      int
		)

		err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk)
		if err != nil {
			rows.Close()
			return err
		}
		if name == "render_version" {
			hasCol = true
		}
	}

	err = rows.Err()
	if err != nil {
		rows.Close()
		return err
	}

	if !hasCol {
		_, err := tx.Exec("ALTER TABLE posts ADD COLUMN render_version INTEGER NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at)")
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", schemaVersion))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func Open(cfg config.Config) (*sql.DB, error) {
	err := os.MkdirAll(cfg.DataDir, 0o755)
	if err != nil {
		return nil, err
	}

	DSN := "file:" + cfg.DBPath + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)&_txlock=immediate"
	db, err := sql.Open("sqlite", DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.Exec(schemaSQL)
	if err != nil {
		db.Close()
		return nil, err
	}

	err = migrate(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	err = checkPragmas(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
