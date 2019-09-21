package fixtures

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-txdb"
	"github.com/stretchr/testify/assert"

	"tgj-bot/th"
)

type PostgreSQLFixture struct {
	T *testing.T
	DB *sql.DB
}

func (fx *PostgreSQLFixture) Finish() {
	assert.NoError(fx.T, fx.DB.Close())
}

func New(t *testing.T,driver, connURL string) *PostgreSQLFixture {
	txDriver := initTxDBDriver(driver, connURL)

	db, err := sql.Open(txDriver, connURL)
	assert.NoError(t, err)
	fx := &PostgreSQLFixture{
		DB: db,
		T: t,
	}

	return fx
}

func initTxDBDriver(driver, dsn string) (txDbEngine string) {
	txDbEngine = th.Stringn(20)
	txdb.Register(txDbEngine, driver, dsn)
	return
}
