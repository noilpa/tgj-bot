package database

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

const (
	driver = "sqlite3"
	dsn    = "./tgj_test.db"
)

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

func TestClient_Dummy(t *testing.T) {
	f := newFixture(t)
	f.finish()
}
