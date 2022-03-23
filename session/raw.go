package session

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chenquan/zap-plus/log"
	"github.com/go-venus/venus/clause"
	"go.uber.org/zap"
)

func (d *DB[T]) Raw(sql string, values ...any) *DB[T] {
	d.sql.WriteString(sql)
	d.sql.WriteString(" ")
	d.sqlVars = append(d.sqlVars, values...)
	log.Debug("sql", zap.String("sql", d.sql.String()), zap.Any("param", d.sqlVars))
	fmt.Println(d.sql.String(), d.sqlVars)
	return d
}

func (d *DB[T]) QueryRow() *sql.Row {
	return d.QueryRowContext(context.Background())
}

func (d *DB[T]) QueryRowContext(ctx context.Context) *sql.Row {
	defer d.Clear()
	return d.getDB().QueryRowContext(ctx, d.sql.String(), d.sqlVars...)
}

func (d *DB[T]) QueryRows() (rows *sql.Rows, err error) {
	return d.QueryRowsContext(context.Background())
}

func (d *DB[T]) QueryRowsContext(ctx context.Context) (rows *sql.Rows, err error) {
	defer d.Clear()
	rows, err = d.getDB().QueryContext(ctx, d.sql.String(), d.sqlVars...)
	return
}

func (d *DB[T]) Exec() (result sql.Result, err error) {
	return d.ExecContext(context.Background())
}

func (d *DB[T]) ExecContext(ctx context.Context) (result sql.Result, err error) {
	defer d.Clear()
	result, err = d.getDB().ExecContext(ctx, d.sql.String(), d.sqlVars...)
	return
}

func (d *DB[T]) Clear() {
	d.sql.Reset()
	d.sqlVars = nil
	d.clause = clause.Clause{}
}

func (d *DB[T]) getDB() db {
	if d.tx != nil {
		return d.tx
	}
	return d.db
}
