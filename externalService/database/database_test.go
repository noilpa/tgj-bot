package database

import (
	"database/sql"
	"os"
	"testing"
	"time"

	th "tgj-bot/testHelpers"

	"tgj-bot/models"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

const (
	driver = "sqlite3"
	dsn    = "./tgj_test.db"
)

func TestClient_Dummy(t *testing.T) {
	f := newFixture(t)
	f.finish()
}

type fixture struct {
	Client
	T *testing.T
}

func newFixture(t *testing.T) *fixture {

	db, err := sql.Open(driver, dsn)
	assert.NoError(t, err)
	f := &fixture{Client:
	Client{db: db},
		T: t,
	}
	assert.NoError(t, initSchema(f.db))
	return f
}

func (f *fixture) finish() {
	f.Close()
	assert.NoError(f.T, os.Remove(dsn))
}

func (f *fixture) createUser() models.User {
	return f.createUsersN(1)[0]
}

func (f *fixture) createUsersN(n int) models.Users {
	q := `INSERT INTO users (telegram_id, telegram_username, gitlab_id, jira_id, is_active, role)
		  VALUES (?, ?, ?, ?, ?, ?)`
	us := make([]models.User, n, n)
	for i := 0; i < n; i++ {
		u := models.User{
			UserBrief: models.UserBrief{
				TelegramID:       th.String(),
				TelegramUsername: th.String(),
				Role:             th.Role(),
			},
			GitlabID: th.String(),
			JiraID:   th.String(),
			IsActive: th.Bool(),
		}
		_, err := f.db.Exec(q, u.TelegramID, u.TelegramUsername, u.GitlabID, u.JiraID, u.IsActive, u.Role)
		assert.NoError(f.T, err)
		us[i] = u
	}
	return us
}

func (f *fixture) getUser(tgUsername string) models.User {
	return f.getUsers(tgUsername)[0]
}

func (f *fixture) getUsers(tgUsername ...string) models.Users {
	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role FROM users WHERE telegram_username = ?`
	n := len(tgUsername)
	us := make([]models.User, n, n)
	u := models.User{}
	for i, name := range tgUsername {
		assert.NoError(f.T, f.db.QueryRow(q, name).Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role))
		us[i] = u
	}
	return us
}

func (f *fixture) createMR(authorID int) models.MR {
	return f.createMrs(authorID)[0]
}

func (f *fixture) createMrs(authorIDs ...int) []models.MR {
	q := `INSERT INTO mrs (url, author_id) VALUES (?, ?)`
	n := len(authorIDs)
	mrs := make([]models.MR, n, n)
	for i := 0; i < n; i++ {
		mr := models.MR{
			URL:      th.String(),
			AuthorID: authorIDs[i],
		}
		_, err := f.db.Exec(q, mr.URL, mr.AuthorID)
		assert.NoError(f.T, err)
	}
	return mrs
}

func (f *fixture) getMR(url string) models.MR {
	return f.getMrs(url)[0]
}

func (f *fixture) getMrs(urls ...string) []models.MR {
	q := `SELECT id, url, author_id, is_closed FROM main.mrs WHERE url = ?`
	n := len(urls)
	mrs := make([]models.MR, n, n)
	mr := models.MR{}
	for i, url := range urls {
		assert.NoError(f.T, f.db.QueryRow(q, url).Scan(&mr.ID, &mr.URL, &mr.AuthorID, &mr.IsClosed))
		mrs[i] = mr
	}
	return mrs
}

// map [userID][]mrIDs
func (f *fixture) createReviews(reviews map[int][]int) []models.Review {
	q := `INSERT INTO reviews (mr_id, user_id, updated_at) VALUES (?, ?, ?)`

	rvs := make([]models.Review, 0)
	for u, mrs := range reviews {
		for _, m := range mrs {
			now := time.Now().Unix()
			_, err := f.db.Exec(q, u, m, now)
			assert.NoError(f.T, err)
			rvs = append(rvs, models.Review{
				MrID:      m,
				UserID:    u,
				UpdatedAt: now,
			})
		}
	}
	return rvs
}
