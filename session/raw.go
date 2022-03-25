package session

import (
	"context"
	"database/sql"

	"github.com/go-venus/venus/clause"
)

func (d *DB[T]) Raw(sql string, values ...any) *DB[T] {
	d.Sql.WriteString(sql)
	d.Sql.WriteString(" ")
	d.SqlVars = append(d.SqlVars, values...)
	return d
}

func (d *DB[T]) QueryRow() *sql.Row {
	return d.QueryRowContext(context.Background())
}

func (d *DB[T]) QueryRowContext(ctx context.Context) *sql.Row {
	defer d.Clear()
	if hook, ok := d.RefTable().Model.(BeforeExecute[T]); ok {
		hook.BeforeExecute(ctx, d)
	}

	defer func() {
		if hook, ok := d.RefTable().Model.(AfterExecute[T]); ok {
			hook.AfterExecute(ctx, d)
		}
	}()

	return d.getDB().QueryRowContext(ctx, d.Sql.String(), d.SqlVars...)
}

func (d *DB[T]) QueryRows() (rows *sql.Rows, err error) {
	return d.QueryRowsContext(context.Background())
}

func (d *DB[T]) QueryRowsContext(ctx context.Context) (rows *sql.Rows, err error) {
	defer d.Clear()
	if hook, ok := d.RefTable().Model.(BeforeExecute[T]); ok {
		hook.BeforeExecute(ctx, d)
	}

	defer func() {
		if hook, ok := d.RefTable().Model.(AfterExecute[T]); ok {
			hook.AfterExecute(ctx, d)
		}
	}()

	rows, err = d.getDB().QueryContext(ctx, d.Sql.String(), d.SqlVars...)
	return
}

func (d *DB[T]) Exec() (result sql.Result, err error) {
	return d.ExecContext(context.Background())
}

func (d *DB[T]) ExecContext(ctx context.Context) (result sql.Result, err error) {
	defer d.Clear()
	if hook, ok := d.RefTable().Model.(BeforeExecute[T]); ok {
		hook.BeforeExecute(ctx, d)
	}

	defer func() {
		if hook, ok := d.RefTable().Model.(AfterExecute[T]); ok {
			hook.AfterExecute(ctx, d)
		}
	}()

	result, err = d.getDB().ExecContext(ctx, d.Sql.String(), d.SqlVars...)
	return
}

func (d *DB[T]) Clear() {
	d.Sql.Reset()
	d.SqlVars = nil
	d.Clause = clause.Clause{}
}

func (d *DB[T]) getDB() db {
	if d.tx != nil {
		return d.tx
	}
	return d.db
}
