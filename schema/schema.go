package schema

import (
	"reflect"
	"strings"
	"sync"

	"golang.org/x/sync/singleflight"
)

var (
	singleFlight = singleflight.Group{}
	tableCache   = map[string]*Table{}
	rwTableCache = sync.RWMutex{}
)

type TableName = string

type Table struct {
	Model            interface{}
	TableName        TableName // 表名
	Fields           []*Field  // 字段
	FieldNames       []string  // 字段名(列名)
	StructFieldNames []string
	fieldMap         map[string] /*字段名(列名)*/ *Field
}

func (s *Table) GetField(name string) *Field {
	return s.fieldMap[name]
}

func Parse[T any](model T, d Dialect) *Table {
	modelType := reflect.Indirect(reflect.ValueOf(model)).Type()
	tableName := strings.ToLower(modelType.Name())
	table, _, _ := singleFlight.Do(tableName, func() (interface{}, error) {
		rwTableCache.RLock()
		table, ok := tableCache[tableName]
		rwTableCache.RUnlock()
		if ok {
			return table, nil
		}

		table = &Table{
			Model:     model,
			TableName: tableName,
			fieldMap:  make(map[string]*Field),
		}

		numField := modelType.NumField()
		for i := 0; i < numField; i++ {
			p := modelType.Field(i)
			if !p.Anonymous && p.IsExported() {
				field := &Field{
					StructName: p.Name,
					Table:      table,
					FieldType:  p.Type,
				}

				if _, ok := p.Tag.Lookup("venus"); ok {
					field.Tag = ParseTag(p.Tag, ";")
				}

				var fieldName string
				if name, ok := field.Tag.TagSettings["COLUMN"]; ok {
					fieldName = name
				} else {
					fieldName = strings.ToLower(p.Name)
				}

				field.Name = fieldName
				table.Fields = append(table.Fields, field)
				table.FieldNames = append(table.FieldNames, fieldName)
				table.StructFieldNames = append(table.StructFieldNames, p.Name)
				table.fieldMap[fieldName] = field

			}
		}
		rwTableCache.RLock()
		tableCache[tableName] = table
		rwTableCache.RUnlock()

		return table, nil

	})

	return table.(*Table)
}

func (s *Table) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))

	var fieldValues []interface{}
	for _, field := range s.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.StructName).Interface())
	}

	return fieldValues
}
