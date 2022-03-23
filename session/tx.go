package session

import (
	"context"
	"database/sql"
)

func (d *DB[T]) cloneDB() *DB[T] {
	return &DB[T]{
		model:    d.model,
		db:       d.db,
		sql:      d.sql,
		sqlVars:  d.sqlVars,
		dialect:  d.dialect,
		refTable: d.refTable,
		clause:   d.clause,
	}
}
func (s *Session[T]) Begin() (tx *Tx[T], err error) {
	return s.BeginTx(context.Background())
}

func (s *Session[T]) BeginTx(ctx context.Context) (tx *Tx[T], err error) {
	var beginTx *sql.Tx
	beginTx, err = s.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	db := s.cloneDB()
	db.tx = beginTx
	return &Tx[T]{
		DB: db,
	}, err
}

func (t *Tx[T]) Commit() (err error) {
	err = t.tx.Commit()
	return
}

func (t *Tx[T]) Rollback() (err error) {
	err = t.tx.Rollback()
	return
}
