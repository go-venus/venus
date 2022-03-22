package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-venus/venus/clause"
	"github.com/go-venus/venus/schema"
)

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
	d.refTable = schema.Parse(d.model, d.dialect)
	return &Session[T]{
		DB: d,
	}
}

func (s *DB[T]) Insert(values ...*T) (int64, error) {
	return s.InsertContext(context.Background(), values...)
}

func (s *DB[T]) InsertContext(ctx context.Context, values ...*T) (int64, error) {
	recordValues := make([]interface{}, 0)
	table := s.RefTable()
	for _, value := range values {
		s.clause.Set(clause.Insert, table.TableName, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))
	}

	s.clause.Set(clause.Values, recordValues...)
	sqlStr, vars := s.clause.Build(clause.Insert, clause.Values)
	result, err := s.Raw(sqlStr, vars...).ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (s *DB[T]) Update(record map[string]interface{}) (int64, error) {
	if beforeUpdate, ok := s.RefTable().Model.(BeforeUpdate[T]); ok {
		err := beforeUpdate.BeforeUpdate(s)
		if err != nil {
			return 0, err
		}
	}

	s.clause.Set(clause.Update, s.RefTable().TableName, record)
	sqlStr, vars := s.clause.Build(clause.Update, clause.Where)

	result, err := s.Raw(sqlStr, vars...).Exec()
	if err != nil {
		return 0, err
	}
	if afterUpdate, ok := s.RefTable().Model.(AfterUpdate[T]); ok {
		if err := afterUpdate.AfterUpdate(s); err != nil {
			rowsAffected, _ := result.RowsAffected()
			return rowsAffected, err
		}
	}

	return result.RowsAffected()
}

func (s *DB[T]) Delete() (int64, error) {
	if afterDelete, ok := s.RefTable().Model.(AfterDelete[T]); ok {
		if err := afterDelete.AfterDelete(s); err != nil {
			return 0, err
		}
	}

	s.clause.Set(clause.Delete, s.RefTable().TableName)
	sqlStr, vars := s.clause.Build(clause.Delete, clause.Where)
	result, err := s.Raw(sqlStr, vars...).Exec()
	if err != nil {
		return 0, err
	}
	if beforeDelete, ok := s.RefTable().Model.(BeforeDelete[T]); ok {
		if err := beforeDelete.BeforeDelete(s); err != nil {
			rowsAffected, _ := result.RowsAffected()
			return rowsAffected, err
		}
	}

	return result.RowsAffected()
}

func (s *DB[T]) Select(values *[]T) (err error) {
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()

	table := s.RefTable()
	s.clause.Set(clause.Select, table.TableName, table.FieldNames)
	sqlStr, vars := s.clause.Build(clause.Select, clause.Where, clause.OrderBy, clause.Limit)
	rows, err := s.Raw(sqlStr, vars...).QueryRows()

	if err != nil {
		return
	}

	// result set
	for rows.Next() {
		dest := reflect.New(destType).Elem()

		var fieldValues []interface{}
		for _, name := range table.StructFieldNames {
			fieldValues = append(fieldValues, dest.FieldByName(name).Addr().Interface())
		}

		if err = rows.Scan(fieldValues...); err != nil {
			return
		}

		destSlice.Set(reflect.Append(destSlice, dest))
	}

	return rows.Close()
}

func (s *DB[T]) Count() (n int64, err error) {
	s.clause.Set(clause.Count, s.RefTable().TableName)
	sqlStr, vars := s.clause.Build(clause.Count, clause.Where)
	row := s.Raw(sqlStr, vars...).QueryRow()
	if err = row.Scan(&n); err != nil {
		return
	}
	return
}

func (s *DB[T]) Limit(num int) *DB[T] {
	s.clause.Set(clause.Limit, num)
	return s
}

func (s *DB[T]) Where(desc string, args ...interface{}) *DB[T] {
	var vars []interface{}
	s.clause.Set(clause.Where, append(append(vars, desc), args...)...)
	return s
}

func (s *DB[T]) OrderBy(desc string) *DB[T] {
	s.clause.Set(clause.OrderBy, desc)
	return s
}

func (s *DB[T]) First(value *T) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Select(destSlice.Addr().Interface().(*[]T)); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	dest.Set(destSlice.Index(0))
	return nil
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

func (s *Tx[T]) NotTransaction(txFn func(db *DB[T]) error) (err error) {
	return txFn(s.cloneDB())
}
