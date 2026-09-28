package database

import (
	"database/sql"
	"errors"

	_ "modernc.org/sqlite"
)

var ErrUserExists = errors.New("username already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrRoomFull = errors.New("chat room is full")

const maxUsers = 2

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		username 			TEXT PRIMARY KEY,
		password_hash	TEXT NOT NULL,
		created_at		DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`)
	return err
}

func UserCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func CreateUser(db *sql.DB, username, passwordHash string) error {
	count, err := UserCount(db)
	if err != nil {
		return err
	}
	if count >= maxUsers {
		return ErrRoomFull
	}

	_, err = db.Exec(
		`INSERT INTO users (username, password_hash) VALUES (?, ?)`,
		username, passwordHash,
	)
	if err != nil {
		return ErrUserExists
	}
	return nil
}

func GetUserPasswordHash(db *sql.DB, username string) (string, error) {
	var hash string

	err := db.QueryRow(
		`SELECT password_hash FROM users WHERE username = ?`,
		username,
	).Scan(&hash)
	if err == sql.ErrNoRows {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}
	return hash, nil

}
