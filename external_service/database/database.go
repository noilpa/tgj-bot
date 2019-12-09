package database

import (
	"database/sql"
	"fmt"

	ce "tgj-bot/custom_errors"
	"tgj-bot/models"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
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
	GetAllMRs() ([]models.MR, error)
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
	DriverName    string `json:"driver"`
	Host          string `json:"host"`
	Port          string `json:"port"`
	User          string `json:"user"`
	Pass          string `json:"pass"`
	DBName        string `json:"dbname"`
	MigrationsDir string `json:"migrations_dir"`
}

func (c *DbConfig) DSN() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", c.User, c.Pass, c.DBName)
}

func (c *DbConfig) MigrationDSN() string {
	return fmt.Sprintf("%s://%s:%s@%s/%s?sslmode=disable", c.DriverName, c.User, c.Pass, c.Host, c.DBName)
}

type Client struct {
	db *sql.DB
}

func RunDB(cfg DbConfig) (dbClient Client, err error) {
	dbClient.db, err = sql.Open(cfg.DriverName, cfg.DSN())
	if err != nil {
		err = ce.WrapWithLog(err, "DB client err")
		return
	}
	if err = dbClient.db.Ping(); err != nil {
		err = ce.WrapWithLog(err, "DB ping err")
		return
	}
	if err = runMigrations(cfg.MigrationDSN(), cfg.MigrationsDir); err != nil {
		return
	}

	return
}

func runMigrations(dsn, path string) error {
	m, err := migrate.New("file://"+path, dsn)
	if err != nil {
		return errors.Wrap(err, "migrations failed")
	}
	err = m.Up()
	if err != migrate.ErrNoChange {
		return errors.Wrap(err, "migrations up failed")
	}
	return nil
}

func (c *Client) Close() {
	c.db.Close()
}
