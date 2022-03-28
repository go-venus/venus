package session

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"

	"github.com/go-venus/venus/clause"
	"github.com/go-venus/venus/dialect"
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
		DestType reflect.Value
		db       *sql.DB
		tx       *sql.Tx
		Sql      strings.Builder
		SqlVars  []any
		dialect  dialect.Dialect
		refTable *schema.Table
		Clause   clause.Clause
	}
	Session[T any] struct {
		*DB[T]
	}

	Tx[T any] struct {
		*DB[T]
	}
)

func New[T any](db *sql.DB, dialect dialect.Dialect) *Session[T] {
	d := &DB[T]{db: db, dialect: dialect}
	d.DestType = reflect.Indirect(reflect.ValueOf(d.model))
	d.refTable = schema.Parse(d.model)
	return &Session[T]{
		DB: d,
	}
}

func (d *DB[T]) Insert(values ...T) (int64, error) {
	return d.InsertContext(context.Background(), values...)
}

func (d *DB[T]) InsertContext(ctx context.Context, values ...T) (rowsAffected int64, err error) {
	table := d.RefTable()
	if afterInsert, ok := table.Model.(AfterInsert[T]); ok {
		if err = afterInsert.AfterInsert(ctx, d); err != nil {
			return
		}
	}

	recordValues := make([]interface{}, 0)
	for _, value := range values {
		d.Clause.Set(clause.Insert, table.TableName, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))
	}

	d.Clause.Set(clause.Values, recordValues...)
	sqlStr, vars := d.Clause.Build(clause.Insert, clause.Values)
	result, err := d.Raw(sqlStr, vars...).ExecContext(ctx)
	if err != nil {
		return
	}

	if afterInsert, ok := table.Model.(AfterInsert[T]); ok {
		if err = afterInsert.AfterInsert(ctx, d); err != nil {
			rowsAffected, _ = result.RowsAffected()
			return
		}
	}

	return result.RowsAffected()
}

func (d *DB[T]) Delete() (int64, error) {
	return d.DeleteContext(context.Background())
}

func (d *DB[T]) DeleteContext(ctx context.Context) (rowsAffected int64, err error) {
	table := d.RefTable()

	if afterDelete, ok := table.Model.(AfterDelete[T]); ok {
		if err = afterDelete.AfterDelete(ctx, d); err != nil {
			return
		}
	}

	d.Clause.Set(clause.Delete, table.TableName)
	sqlStr, vars := d.Clause.Build(clause.Delete, clause.Where)
	result, err := d.Raw(sqlStr, vars...).ExecContext(ctx)
	if err != nil {
		return
	}

	if beforeDelete, ok := table.Model.(BeforeDelete[T]); ok {
		if err = beforeDelete.BeforeDelete(ctx, d); err != nil {
			rowsAffected, _ = result.RowsAffected()
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
	if beforeQuery, ok := table.Model.(BeforeQuery[T]); ok {
		if err = beforeQuery.BeforeQuery(ctx, d); err != nil {
			return
		}
	}

	d.Clause.Set(clause.Select, table.TableName, table.FieldNames)
	sqlStr, vars := d.Clause.Build(clause.Select, clause.Where, clause.OrderBy, clause.Limit)
	rows, err := d.Raw(sqlStr, vars...).QueryRowsContext(ctx)

	if err != nil {
		return
	}
	defer func() {
		err = rows.Close()
	}()

	// results set
	for rows.Next() {
		dest := reflect.New(d.DestType.Type()).Elem()

		fieldValues := make([]interface{}, len(table.StructFieldNames))
		for i, name := range table.StructFieldNames {
			fieldValues[i] = dest.FieldByName(name).Addr().Interface()
		}

		if err = rows.Scan(fieldValues...); err != nil {
			return
		}

		t := dest.Interface().(T)
		results = append(results, t)
	}

	if afterQuery, ok := table.Model.(AfterQuery[T]); ok {
		if err = afterQuery.AfterQuery(ctx, d); err != nil {
			return
		}
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
	table := d.RefTable()
	if beforeQuery, ok := table.Model.(BeforeQuery[T]); ok {
		if err = beforeQuery.BeforeQuery(ctx, d); err != nil {
			return
		}
	}

	d.Clause.Set(clause.Count, d.RefTable().TableName)
	sqlStr, vars := d.Clause.Build(clause.Count, clause.Where)
	row := d.Raw(sqlStr, vars...).QueryRowContext(ctx)
	if err = row.Scan(&n); err != nil {
		return
	}

	if beforeDelete, ok := table.Model.(BeforeDelete[T]); ok {
		if err = beforeDelete.BeforeDelete(ctx, d); err != nil {
			return
		}
	}

	return
}

func (d *DB[T]) Update(record map[string]interface{}) (int64, error) {
	return d.UpdateContext(context.Background(), record)
}

func (d *DB[T]) UpdateContext(ctx context.Context, record map[string]interface{}) (rowsAffected int64, err error) {
	table := d.RefTable()
	if beforeUpdate, ok := table.Model.(BeforeUpdate[T]); ok {
		if err = beforeUpdate.BeforeUpdate(ctx, d); err != nil {
			return
		}
	}

	d.Clause.Set(clause.Update, table.TableName, record)
	sqlStr, vars := d.Clause.Build(clause.Update, clause.Where)

	result, err := d.Raw(sqlStr, vars...).ExecContext(ctx)
	if err != nil {
		return
	}

	if afterUpdate, ok := table.Model.(AfterUpdate[T]); ok {
		if err = afterUpdate.AfterUpdate(ctx, d); err != nil {
			rowsAffected, _ = result.RowsAffected()
			return
		}
	}

	return result.RowsAffected()
}

func (d *DB[T]) Limit(num int) *DB[T] {
	d.Clause.Set(clause.Limit, num)
	return d
}

func (d *DB[T]) Where(desc string, args ...interface{}) *DB[T] {
	var vars []interface{}
	d.Clause.Set(clause.Where, append(append(vars, desc), args...)...)
	return d
}

func (d *DB[T]) OrderBy(desc string) *DB[T] {
	d.Clause.Set(clause.OrderBy, desc)
	return d
}
