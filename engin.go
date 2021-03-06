package venus

import (
	"database/sql"

	"github.com/go-venus/venus/dialect"
	"github.com/go-venus/venus/session"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func Open(config *Config) (e *Engine, err error) {
	db, err := sql.Open(config.Driver, config.Source)
	if err != nil {
		return
	}
	// Send a ping to make sure the database connection is alive.
	if err = db.Ping(); err != nil {
		return
	}
	// make sure the specific dialect exists
	dial, err := dialect.GetDialect(config.Driver)
	if err != nil {
		return
	}

	e = &Engine{db: db, dialect: dial}
	return
}

func NewSession[T any](e *Engine) *session.Session[T] {
	return session.New[T](e.db, e.dialect)
}
