package session

import (
	"context"
	"database/sql"
	"fmt"
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
func (s *Session[T]) Transaction(txFn func(tx *Tx[T]) error) (err error) {
	var tx *Tx[T]
	if tx, err = s.Begin(); err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			rollbackErr := tx.Rollback() // err is non-nil; don't change it
			if rollbackErr != nil {
				err = fmt.Errorf("excute err: %w, tx error: %s", err, rollbackErr.Error())
			}
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
		}
	}()

	return txFn(tx)
}

func (t *Tx[T]) NotTransaction(txFn func(db *DB[T]) error) (err error) {
	return txFn(t.cloneDB())
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
