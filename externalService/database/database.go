package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"tgj-bot/models"
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
					  role TEXT);`

	createMrs := `CREATE TABLE IF NOT EXISTS mrs (
  				    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, 
					name TEXT);`

	createReviews := `CREATE TABLE IF NOT EXISTS reviews (
					    mr_id INTEGER NOT NULL,
					    user_id INTEGER NOT NULL,
					    is_approved INTEGER,
					    is_commented INTEGER,
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

func (c *Client) SaveUser(user models.User) (err error) {
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

func(c *Client) ChangeIsActiveUser(telegramID string, isActive bool) (err error) {
	q := `UPDATE users SET is_active = ? WHERE telegram_id = ?`
	res, err := c.Db.Exec(q, isActive, telegramID)
	if err != nil {
		return errors.New(fmt.Sprint("change is active user error: %v", err))
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("change is active user error: no affected rows")
	}
	return
}

func (c *Client) GetUsersWithPayload(telegramID string) (ups models.UsersPayload, err error) {
	q := `SELECT id, 
       			 telegram_id, 
       			 role, 
                 (SELECT count(*) 
                  FROM reviews r 
                  WHERE r.user_id = id
                    AND r.is_approved = FALSE) AS payload
          FROM users
		  WHERE telegram_id != ?
			AND is_active = TRUE
		  ORDER BY role, payload;`

	rows, err := c.Db.Query(q, telegramID)
	if err != nil {
		return
	}
	defer rows.Close()

	var up models.UserPayload
	for rows.Next() {
		if err = rows.Scan(&up.ID, &up.TelegramID, &up.Role, &up.Payload); err != nil {
			return
		}
		ups = append(ups, up)
	}
	return
}