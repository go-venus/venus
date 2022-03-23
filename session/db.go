package session

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"

	"github.com/go-venus/venus/clause"
	"github.com/go-venus/venus/schema"
)

var ErrNotFound = errors.New("not found")

type (
	db interface {
		Query(query string, args ...any) (*sql.Rows, error)
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
		QueryRow(query string, args ...any) *sql.Row
		QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
		Exec(query string, args ...any) (sql.Result, error)
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	}

	DB[T any] struct {
		model    T
		destType reflect.Value
		db       *sql.DB
		tx       *sql.Tx
		sql      strings.Builder
		sqlVars  []any
		dialect  schema.Dialect
		refTable *schema.Table
		clause   clause.Clause
	}
	Session[T any] struct {
		*DB[T]
	}

	Tx[T any] struct {
		*DB[T]
	}
)

func New[T any](db *sql.DB, dialect schema.Dialect) *Session[T] {
	d := &DB[T]{db: db, dialect: dialect}
	d.destType = reflect.Indirect(reflect.ValueOf(d.model))
	d.refTable = schema.Parse(d.model, d.dialect)
	return &Session[T]{
		DB: d,
	}
}

func (d *DB[T]) Insert(values ...*T) (int64, error) {
	return d.InsertContext(context.Background(), values...)
}

func (d *DB[T]) InsertContext(ctx context.Context, values ...*T) (int64, error) {
	recordValues := make([]interface{}, 0)
	table := d.RefTable()
	for _, value := range values {
		d.clause.Set(clause.Insert, table.TableName, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))
	}

	d.clause.Set(clause.Values, recordValues...)
	sqlStr, vars := d.clause.Build(clause.Insert, clause.Values)
	result, err := d.Raw(sqlStr, vars...).ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (d *DB[T]) Delete() (int64, error) {
	return d.DeleteContext(context.Background())
}

func (d *DB[T]) DeleteContext(ctx context.Context) (int64, error) {
	if afterDelete, ok := d.RefTable().Model.(AfterDelete[T]); ok {
		if err := afterDelete.AfterDelete(d); err != nil {
			return 0, err
		}
	}

	d.clause.Set(clause.Delete, d.RefTable().TableName)
	sqlStr, vars := d.clause.Build(clause.Delete, clause.Where)
	result, err := d.Raw(sqlStr, vars...).ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	if beforeDelete, ok := d.RefTable().Model.(BeforeDelete[T]); ok {
		if err := beforeDelete.BeforeDelete(d); err != nil {
			rowsAffected, _ := result.RowsAffected()
			return rowsAffected, err
		}
	}

	return result.RowsAffected()
}

func (d *DB[T]) Select() (results []T, err error) {
	return d.SelectContext(context.Background())
}

func (d *DB[T]) SelectContext(ctx context.Context) (results []T, err error) {
	table := d.RefTable()
	d.clause.Set(clause.Select, table.TableName, table.FieldNames)
	sqlStr, vars := d.clause.Build(clause.Select, clause.Where, clause.OrderBy, clause.Limit)
	rows, err := d.Raw(sqlStr, vars...).QueryRowsContext(ctx)

	if err != nil {
		return
	}
	defer func() {
		err = rows.Close()
	}()

	// results set
	for rows.Next() {
		dest := reflect.New(d.destType.Type()).Elem()

		var fieldValues []interface{}
		for _, name := range table.StructFieldNames {
			fieldValues = append(fieldValues, dest.FieldByName(name).Addr().Interface())
		}

		if err = rows.Scan(fieldValues...); err != nil {
			return
		}

		t := dest.Interface().(T)
		results = append(results, t)
	}

	return
}

func (d *DB[T]) First() (result T, err error) {
	return d.FirstContext(context.Background())
}

func (d *DB[T]) FirstContext(ctx context.Context) (result T, err error) {
	results, err := d.Limit(1).SelectContext(ctx)
	if err != nil {
		return
	}

	if len(results) == 0 {
		err = ErrNotFound
		return
	}

	return results[0], nil
}

func (d *DB[T]) Count() (n int64, err error) {
	return d.CountContext(context.Background())
}

func (d *DB[T]) CountContext(ctx context.Context) (n int64, err error) {
	d.clause.Set(clause.Count, d.RefTable().TableName)
	sqlStr, vars := d.clause.Build(clause.Count, clause.Where)
	row := d.Raw(sqlStr, vars...).QueryRowContext(ctx)
	if err = row.Scan(&n); err != nil {
		return
	}
	return
}

func (d *DB[T]) Update(record map[string]interface{}) (int64, error) {
	return d.UpdateContext(context.Background(), record)
}

func (d *DB[T]) UpdateContext(ctx context.Context, record map[string]interface{}) (int64, error) {
	if beforeUpdate, ok := d.RefTable().Model.(BeforeUpdate[T]); ok {
		err := beforeUpdate.BeforeUpdate(d)
		if err != nil {
			return 0, err
		}
	}

	d.clause.Set(clause.Update, d.RefTable().TableName, record)
	sqlStr, vars := d.clause.Build(clause.Update, clause.Where)

	result, err := d.Raw(sqlStr, vars...).ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	if afterUpdate, ok := d.RefTable().Model.(AfterUpdate[T]); ok {
		if err := afterUpdate.AfterUpdate(d); err != nil {
			rowsAffected, _ := result.RowsAffected()
			return rowsAffected, err
		}
	}

	return result.RowsAffected()
}

func (d *DB[T]) Limit(num int) *DB[T] {
	d.clause.Set(clause.Limit, num)
	return d
}

func (d *DB[T]) Where(desc string, args ...interface{}) *DB[T] {
	var vars []interface{}
	d.clause.Set(clause.Where, append(append(vars, desc), args...)...)
	return d
}

func (d *DB[T]) OrderBy(desc string) *DB[T] {
	d.clause.Set(clause.OrderBy, desc)
	return d
}
