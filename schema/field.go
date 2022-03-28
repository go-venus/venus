package schema

import (
	"reflect"
)

type DataType string

type Field struct {
	StructName  string
	Name        string
	FieldType   reflect.Type
	StructField reflect.StructField
	Tag         *Tag
	Table       *Table
}
