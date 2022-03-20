package venus

import (
	"database/sql"

	"github.com/go-venus/venus/schema"
	"github.com/go-venus/venus/session"
)

type Engine struct {
	db      *sql.DB
	dialect schema.Dialect
}

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		return
	}
	// Send a ping to make sure the database connection is alive.
	if err = db.Ping(); err != nil {
		return
	}
	// make sure the specific dialect exists
	dial, err := schema.GetDialect(driver)
	if err != nil {
		return
	}

	e = &Engine{db: db, dialect: dial}
	return
}

func NewSession[T any](e *Engine) *session.Session[T] {
	return session.New[T](e.db, e.dialect)
}
