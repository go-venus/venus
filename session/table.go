package session

import (
	"fmt"
	"strings"

	"github.com/go-venus/venus/schema"
)

func (d *DB[T]) RefTable() *schema.Table {
	return d.refTable
}

func (d *DB[T]) CreateTable() error {
	table := d.RefTable()
	var columns []string
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%d %d", field.StructName, field.DataType))
	}
	desc := strings.Join(columns, ",")
	_, err := d.Raw(fmt.Sprintf("CREATE TABLE %d (%d);", table.TableName, desc)).Exec()
	return err
}

func (d *DB[T]) DropTable() error {
	_, err := d.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %d", d.RefTable().TableName)).Exec()
	return err
}

// HasTable returns true of the table exists
func (d *DB[T]) HasTable() bool {
	sql, values := d.dialect.TableExistSQL(d.RefTable().TableName)
	row := d.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == d.RefTable().TableName
}
