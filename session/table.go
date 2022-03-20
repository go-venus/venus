package session

import (
	"fmt"
	"strings"

	"github.com/go-venus/venus/schema"
)

func (s *DB[T]) RefTable() *schema.Table {
	return s.refTable
}

func (s *DB[T]) CreateTable() error {
	table := s.RefTable()
	var columns []string
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s", field.StructName, field.DataType))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.TableName, desc)).Exec()
	return err
}

func (s *DB[T]) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().TableName)).Exec()
	return err
}

// HasTable returns true of the table exists
func (s *DB[T]) HasTable() bool {
	sql, values := s.dialect.TableExistSQL(s.RefTable().TableName)
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == s.RefTable().TableName
}
