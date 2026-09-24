package database

import "database/sql"

func InitializeSchema(db *sql.DB) error {
	const usersTable = `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	const messagesTable = `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

		FOREIGN KEY (sender_id)
			REFERENCES users(id)
	);
	`

	if _, err := db.Exec(usersTable); err != nil {
		return err
	}

	if _, err := db.Exec(messagesTable); err != nil {
		return err
	}

	return nil
}
