package database

import (
	"database/sql"

	ce "tgj-bot/custom_errors"
	"tgj-bot/models"

	_ "github.com/lib/pq"
)

type Database interface {
	Close()
}

type UserRepository interface {
	SaveUser(u models.User) (int, error)
	ChangeIsActiveUser(telegramUsername string, isActive bool) (err error)
	GetUsersWithPayload(exceptTelegramID string) (ups models.UsersPayload, err error)
	GetUserByTgUsername(tgUname string) (u models.User, err error)
	GetUserByGitlabID(id interface{}) (u models.User, err error)
	GetUsersByMrID(id int) (us []models.UserBrief, err error)
	GetUsersByMrURL(url string) (us []models.UserBrief, err error)
	GetUserForReallocateMR(u models.UserBrief, mID int) (up models.UserPayload, err error)
	GetActiveUsers() (us models.UserList, err error)
}

type MergeRequestRepository interface {
	SaveMR(mr models.MR) (models.MR, error)
	GetOpenedMRs() (mrs []models.MR, err error)
	CloseMRs() error
	GetMrByID(id int) (mr models.MR, err error)
	GetMRbyURL(url string) (mr models.MR, err error)
}

type ReviewRepository interface {
	SaveReview(r models.Review) (err error)
	UpdateReviewApprove(r models.Review) error
	UpdateReviewComment(r models.Review) (err error)
	GetReviewMRsByUserID(uID int) (ids []int, err error)
	DeleteReview(r models.Review) (err error)
	GetOpenedReviewsByUserID(uID int) (rs []models.Review, err error)
}

type DbConfig struct {
	DriverName string `json:"driver"`
	DSN        string `json:"dsn"`
}

type Client struct {
	db *sql.DB
}

func RunDB(cfg DbConfig) (dbClient Client, err error) {
	dbClient.db, err = sql.Open(cfg.DriverName, cfg.DSN)
	if err != nil {
		err = ce.WrapWithLog(err, "DB client err")
		return
	}
	if err = dbClient.db.Ping(); err != nil {
		err = ce.WrapWithLog(err, "DB ping err")
		return
	}
	if err = dbClient.initSchema(); err != nil {
		return
	}

	return
}

func (c *Client) initSchema() (err error) {
	setSearchPath := `SET search_path TO public;`

	createUsers := `CREATE TABLE IF NOT EXISTS users (
					  id SERIAL PRIMARY KEY, 
					  telegram_id TEXT UNIQUE,
					  telegram_username TEXT UNIQUE,
					  gitlab_id TEXT UNIQUE, 
					  jira_id TEXT, 
					  is_active BOOLEAN, 
					  role TEXT);`

	createMrs := `CREATE TABLE IF NOT EXISTS mrs (
  				    id SERIAL PRIMARY KEY, 
					url TEXT UNIQUE,
					author_id INTEGER NOT NULL,
					is_closed BOOLEAN DEFAULT FALSE,
					FOREIGN KEY(author_id) REFERENCES users(id));`

	createReviews := `CREATE TABLE IF NOT EXISTS reviews (
					    mr_id INTEGER NOT NULL,
					    user_id INTEGER NOT NULL,
					    is_approved BOOLEAN DEFAULT FALSE,
					    is_commented BOOLEAN DEFAULT FALSE,
					    updated_at BIGINT,
					    PRIMARY KEY (mr_id, user_id),
					    FOREIGN KEY(mr_id) REFERENCES mrs(id),
					    FOREIGN KEY(user_id) REFERENCES users(id));`

	_, err = c.db.Exec(setSearchPath + createUsers + createMrs + createReviews)
	if err != nil {
		err = ce.WrapWithLog(err, "create tables")
		return
	}
	return
}

func (c *Client) Close() {
	c.db.Close()
}
