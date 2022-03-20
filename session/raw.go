package session

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chenquan/zap-plus/log"
	"github.com/go-venus/venus/clause"
	"go.uber.org/zap"
)

func (s *DB[T]) Raw(sql string, values ...any) *DB[T] {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	log.Debug("sql", zap.String("sql", s.sql.String()), zap.Any("param", s.sqlVars))
	fmt.Println(s.sql.String(), s.sqlVars)
	return s
}

func (s *DB[T]) QueryRow() *sql.Row {
	return s.QueryRowContext(context.Background())
}

func (s *DB[T]) QueryRowContext(ctx context.Context) *sql.Row {
	defer s.Clear()
	return s.getDB().QueryRowContext(ctx, s.sql.String(), s.sqlVars...)
}

func (s *DB[T]) QueryRows() (rows *sql.Rows, err error) {
	return s.QueryRowsContext(context.Background())
}

func (s *DB[T]) QueryRowsContext(ctx context.Context) (rows *sql.Rows, err error) {
	defer s.Clear()
	rows, err = s.getDB().QueryContext(ctx, s.sql.String(), s.sqlVars...)
	return
}

func (s *DB[T]) Exec() (result sql.Result, err error) {
	return s.ExecContext(context.Background())
}

func (s *DB[T]) ExecContext(ctx context.Context) (result sql.Result, err error) {
	defer s.Clear()
	result, err = s.getDB().ExecContext(ctx, s.sql.String(), s.sqlVars...)
	return
}

func (s *DB[T]) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

func (s *DB[T]) getDB() db {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}
