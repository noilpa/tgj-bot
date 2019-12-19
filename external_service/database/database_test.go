package database

import (
	"testing"
	"time"

	"tgj-bot/fixtures"
	"tgj-bot/models"
	"tgj-bot/th"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var conf = DbConfig{
	DriverName:    TestDriverName,
	Host:          TestHost,
	Port:          TestPort,
	User:          TestUser,
	Pass:          TestPass,
	DBName:        TestDBName,
	MigrationsDir: TestMigrationsDir,
}

func TestClient_Dummy(t *testing.T) {
	f := newFixture(t)
	f.finish()
}

type fixture struct {
	Client
	T *testing.T
}

func newFixture(t *testing.T) *fixture {
	db := fixtures.New(t, conf.DriverName, conf.MigrationDSN()).DB
	require.NoError(t, db.Ping())
	f := &fixture{
		Client: Client{
			db: db,
		},
		T: t,
	}

	require.NoError(t, runMigrations(conf.MigrationDSN(), conf.MigrationsDir))
	return f
}

func (f *fixture) finish() {
	f.Close()
}

func (f *fixture) createUser() models.User {
	return f.createUsersN(1)[0]
}

func (f *fixture) createUsersN(n int) []models.User {
	q := `INSERT INTO users (telegram_id, telegram_username, gitlab_id, jira_id, is_active, role, gitlab_name)
		  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	us := make([]models.User, n, n)
	for i := 0; i < n; i++ {
		u := models.User{
			UserBrief: models.UserBrief{
				TelegramID:       th.String(),
				TelegramUsername: th.String(),
				Role:             models.Developer,
				GitlabID:         th.Int(),
				GitlabName:       th.String(),
			},
			JiraID:   th.String(),
			IsActive: true,
		}
		assert.NoError(f.T, f.db.QueryRow(q, u.TelegramID, u.TelegramUsername, u.GitlabID, u.JiraID, u.IsActive, u.Role, u.GitlabName).Scan(&u.ID))
		us[i] = u
	}
	return us
}

func (f *fixture) getUser(tgUsername string) models.User {
	return f.getUsers(tgUsername)[0]
}

func (f *fixture) getUsers(tgUsername ...string) models.UserList {
	q := `SELECT id, telegram_id, telegram_username, gitlab_id, jira_id, is_active, role, gitlab_name FROM users WHERE telegram_username = $1`
	n := len(tgUsername)
	us := make([]models.User, n, n)
	u := models.User{}
	for i, name := range tgUsername {
		assert.NoError(f.T, f.db.QueryRow(q, name).Scan(&u.ID, &u.TelegramID, &u.TelegramUsername, &u.GitlabID, &u.JiraID, &u.IsActive, &u.Role, &u.GitlabName))
		us[i] = u
	}
	return us
}

func (f *fixture) createMR(authorID int) models.MR {
	return f.createMRs(authorID, 1)[0]
}

func (f *fixture) createMRs(authorID, n int) []models.MR {
	q := `INSERT INTO mrs (url, author_id, need_jira_update) VALUES ($1,$2,$3) RETURNING id`
	mrs := make([]models.MR, n, n)
	for i := 0; i < n; i++ {
		mr := models.MR{
			URL:            th.String(),
			AuthorID:       &authorID,
			NeedJiraUpdate: true,
			NeedQANotify:   true,
		}
		assert.NoError(f.T, f.db.QueryRow(q, mr.URL, mr.AuthorID, mr.NeedJiraUpdate).Scan(&mr.ID))
		mrs[i] = mr
	}
	return mrs
}

func (f *fixture) closeMR(id int) {
	q := `UPDATE mrs SET is_closed = TRUE WHERE id = $1`
	_, err := f.db.Exec(q, id)
	assert.NoError(f.T, err)
}

func (f *fixture) getMR(url string) models.MR {
	return f.getMRs(url)[0]
}

func (f *fixture) getMRs(urls ...string) []models.MR {
	q := `SELECT id, url, author_id, is_closed FROM mrs WHERE url = $1`
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
	q := `INSERT INTO reviews (mr_id, user_id, updated_at) VALUES ($1, $2, $3)`

	rvs := make([]models.Review, 0)
	for u, mrs := range reviews {
		for _, m := range mrs {
			now := time.Now().Unix()
			_, err := f.db.Exec(q, m, u, now)
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

func (f *fixture) getReviewsByMR(mrID int) (rs []models.Review) {
	q := `SELECT mr_id, user_id, is_approved, is_commented, updated_at FROM reviews WHERE mr_id = $1`
	rows, err := f.db.Query(q, mrID)
	assert.NoError(f.T, err)
	defer rows.Close()

	var r models.Review
	for rows.Next() {
		assert.NoError(f.T, rows.Scan(&r.MrID, &r.UserID, &r.IsApproved, &r.IsCommented, &r.UpdatedAt))
		rs = append(rs, r)
	}
	return
}

func (f *fixture) getReviewsByUser(uID int) (rs []models.Review) {
	q := `SELECT mr_id, user_id, is_approved, is_commented, updated_at FROM reviews WHERE user_id = $1`
	rows, err := f.db.Query(q, uID)
	assert.NoError(f.T, err)
	defer rows.Close()

	var r models.Review
	for rows.Next() {
		assert.NoError(f.T, rows.Scan(&r.MrID, &r.UserID, &r.IsApproved, &r.IsCommented, &r.UpdatedAt))
		rs = append(rs, r)
	}
	return
}

func isContain(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
