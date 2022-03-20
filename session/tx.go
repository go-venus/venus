package session

import (
	"context"
	"database/sql"
)

func (s *DB[T]) cloneDB() *DB[T] {
	return &DB[T]{
		model:    s.model,
		db:       s.db,
		sql:      s.sql,
		sqlVars:  s.sqlVars,
		dialect:  s.dialect,
		refTable: s.refTable,
		clause:   s.clause,
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

func (s *Tx[T]) Commit() (err error) {
	err = s.tx.Commit()
	return
}

func (s *Tx[T]) Rollback() (err error) {
	err = s.tx.Rollback()
	return
}
