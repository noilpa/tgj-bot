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

func (c *Client) SaveMR(mr models.MR) (models.MR, error) {
	q := `INSERT INTO mrs (url, author_id) VALUES (?, ?);
		  SELECT id FROM mrs WHERE url = ?`
	if err := c.db.QueryRow(q, mr.URL, mr.AuthorID).Scan(&mr.ID); err != nil {
		return models.MR{}, err
	}
	return mr, nil
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

//func (c *Client) GetReviewsByMrID(id int) ([]models.Review, error) {
//	q := `SELECT FROM reviews WHERE mr_id = ?`
//	// нужно ли возвращать is_approved = true?
//}

func (c *Client) UpdateReviewApprove(r models.Review) error {
	q := `UPDATE reviews 
			SET is_approved = ?,
				updated_at = ?
		  WHERE user_id = ? 
  			AND mr_id = ?`
	res, err := c.db.Exec(q, r.IsApproved, r.UpdatedAt, r.UserID, r.MrID)
	if err != nil {
		return ce.Wrap(ce.ErrChangeReviewApprove, err.Error())
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ce.Wrap(ce.ErrChangeReviewApprove, "no affected rows")
	}
	return nil
}

func (c *Client) GetOpenedMRs() (mrs []models.MR, err error) {
	q := `SELECT id, url, author_id FROM mrs WHERE is_closed = FALSE`
	rows, err := c.db.Query(q)
	if err != nil {
		return
	}
	defer rows.Close()

	var mr models.MR
	for rows.Next() {
		if err = rows.Scan(&mr.ID, &mr.URL, &mr.AuthorID); err != nil {
			return
		}
		mrs = append(mrs, mr)
	}
	return
}
