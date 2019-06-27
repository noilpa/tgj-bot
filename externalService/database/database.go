package database

import (
	"database/sql"
	"errors"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type DbConfig struct {
	DriverName        string `json:"driver"`
	DSN               string `json:"dsn"`
	IsNeedRemoveOldDB bool   `json:"is_need_remove_old_db"`
}

type Client struct {
	db *sql.DB
}

func RunDB(cfg DbConfig) (dbClient Client, err error) {
	if cfg.IsNeedRemoveOldDB {
		// ignore err if file doesn't exist
		os.Remove(cfg.DSN)
	}
	dbClient.db, err = sql.Open(cfg.DriverName, cfg.DSN)
	if err != nil {
		return
	}
	if err = initSchema(dbClient.db); err != nil {
		return
	}

	return
}

func initSchema(db *sql.DB) (err error) {
	createUsers := `CREATE TABLE IF NOT EXISTS users (
					  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
					  telegram_id TEXT UNIQUE,
					  telegram_username TEXT,
					  gitlab_id TEXT UNIQUE, 
					  jira_id TEXT UNIQUE, 
					  is_active INTEGER, 
					  role TEXT);`

	createMrs := `CREATE TABLE IF NOT EXISTS mrs (
  				    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
					url TEXT UNIQUE,
					author_id INTEGER NOT NULL,
					is_closed INTEGER DEFAULT 0,
					FOREIGN KEY(author_id) REFERENCES users(id));`

	createReviews := `CREATE TABLE IF NOT EXISTS reviews (
					    mr_id INTEGER NOT NULL,
					    user_id INTEGER NOT NULL,
					    is_approved INTEGER DEFAULT 0,
					    is_commented INTEGER DEFAULT 0,
					    updated_at INTEGER,
					    PRIMARY KEY (mr_id, user_id),
					    FOREIGN KEY(mr_id) REFERENCES mrs(id),
					    FOREIGN KEY(user_id) REFERENCES users(id));`

	_, err = db.Exec(createUsers + createMrs + createReviews)
	if err != nil {
		return errors.New("create tables err: " + err.Error())
	}
	return
}

func (c *Client) Close() {
	c.db.Close()
}
