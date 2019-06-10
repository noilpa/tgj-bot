package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"tgj-bot/utils"
)

type DbConfig struct {
	DriverName        string `json:"driver"`
	DSN               string `json:"dsn"`
	IsNeedRemoveOldDB bool   `json:"is_need_remove_old_db"`
}

type Client struct {
	Db *sql.DB
}

func RunDB(cfg DbConfig) (dbClient Client, err error) {
	if cfg.IsNeedRemoveOldDB {
		// ignore err if file doesn't exist
		os.Remove(cfg.DSN)
	}
	dbClient.Db, err = sql.Open(cfg.DriverName, cfg.DSN)
	if err != nil {
		return
	}
	if err = initSchema(dbClient.Db); err != nil {
		return
	}

	return
}

func initSchema(db *sql.DB) (err error) {
	createUsers := `CREATE TABLE IF NOT EXISTS users (
					  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
					  telegram_id TEXT, 
					  gitlab_id TEXT, 
					  jira_id TEXT, 
					  is_active INTEGER, 
					  role TEXT)`

	createMrs := `CREATE TABLE IF NOT EXISTS mrs (
  				    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
					name TEXT)`

	createReviews := `CREATE TABLE IF NOT EXISTS reviews (
					    mr_id INTEGER NOT NULL,
					    user_id INTEGER NOT NULL,
					    status TEXT,
					    PRIMARY KEY (mr_id, user_id),
					    FOREIGN KEY(mr_id) REFERENCES mrs(id),
					    FOREIGN KEY(user_id) REFERENCES users(id))`

	// status:
	// like    -> success ^
	// comment -> pending ^
	// nothing -> init    ^

	// add column updated_at

	_, err = db.Exec(createUsers)
	if err != nil {
		return errors.New("create table 'users' err: " + err.Error())
	}
	_, err = db.Exec(createMrs)
	if err != nil {
		return errors.New("create table 'mrs' err: " + err.Error())
	}
	_, err = db.Exec(createReviews)
	if err != nil {
		return errors.New("create table 'reviews' err: " + err.Error())
	}

	return
}

func (c *Client) SaveUser(user utils.User) (err error) {
	q := `INSERT INTO users (telegram_id, gitlab_id, jira_id, is_active, role)
		  VALUES (?, ?, ?, ?, ?)`
	res, err := c.Db.Exec(q, user.TelegramID, user.GitlabID, user.JiraID, user.IsActive, user.Role)
	if err != nil {
		return errors.New(fmt.Sprint("create new user error: %v", err))
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("create new user error: no affected rows")
	}
	return
}
