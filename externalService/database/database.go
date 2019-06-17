package database

import (
	"database/sql"
	"errors"
	"os"

	ce "tgj-bot/customErrors"
	"tgj-bot/models"

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
					url TEXT UNIQUE);`

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

func (c *Client) SaveUser(user models.User) (err error) {
	q := `INSERT INTO users (telegram_id, telegram_username, gitlab_id, jira_id, is_active, role)
		  VALUES (?, ?, ?, ?, ?, ?)`
	res, err := c.db.Exec(q, user.TelegramID, user.TelegramUsername, user.GitlabID, user.JiraID, user.IsActive, user.Role)
	if err != nil {
		return ce.ErrCreateUser
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ce.Wrap(ce.ErrCreateUser,"no affected rows")
	}
	return
}

func (c *Client) ChangeIsActiveUser(telegramUsername string, isActive bool) (err error) {
	q := `UPDATE users SET is_active = ? WHERE telegram_username = ?`
	res, err := c.db.Exec(q, isActive, telegramUsername)
	if err != nil {
		return ce.Wrap(ce.ErrChangeUserActivity, err.Error())
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ce.Wrap(ce.ErrChangeUserActivity, "no affected rows")
	}
	return
}

func (c *Client) GetUsersWithPayload(telegramID string) (ups models.UsersPayload, err error) {
	q := `SELECT id, 
       			 telegram_id,
       			 telegram_username,
       			 role, 
                 (SELECT count(*) 
                  FROM reviews r 
                  WHERE r.user_id = id
                    AND r.is_approved = FALSE) AS payload
          FROM users
		  WHERE telegram_id != ?
			AND is_active = TRUE
		  ORDER BY role, payload;`

	rows, err := c.db.Query(q, telegramID)
	if err != nil {
		return
	}
	defer rows.Close()

	var up models.UserPayload
	for rows.Next() {
		if err = rows.Scan(&up.ID, &up.TelegramID, &up.TelegramUsername, &up.Role, &up.Payload); err != nil {
			return
		}
		ups = append(ups, up)
	}
	return
}

func (c *Client) GetUsers() (users models.Users, err error) {
	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role 
		  FROM users`

	rows, err := c.db.Query(q)
	if err != nil {
		return
	}
	defer rows.Close()

	var u models.User
	for rows.Next() {
		if err = rows.Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role); err != nil {
			return
		}
		users = append(users, u)
	}
	return
}

func (c *Client) GetUserByTgUsername(id string) (u models.User, err error) {
	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role 
		  FROM users 
          WHERE telegram_username = ?`
	err = c.db.QueryRow(q, id).Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role)
	if err != nil {
		return
	}
	return
}

func (c *Client) SaveMR(url string) (mr models.MR, err error) {
	q := `INSERT INTO mrs (url) VALUES (?);
		  SELECT id FROM mrs WHERE url = ?`
	if err = c.db.QueryRow(q, url, url).Scan(&mr.ID); err != nil {
		return
	}
	mr.URL = url
	return
}

func (c *Client) SaveReview(r models.Review) (err error) {
	q := `INSERT INTO reviews (mr_id, user_id, updated_at) VALUES (?, ?, ?)`
	res, err := c.db.Exec(q, r.MrID, r.UserID, r.UpdatedAt)
	if err != nil {
		return ce.ErrCreateReview
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ce.Wrap(ce.ErrCreateReview, "no affected rows")
	}
	return
}
