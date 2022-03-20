package schema

import (
	"reflect"
)

type DataType string

type Field struct {
	StructName        string
	Name              string
	DataType          DataType
	FieldType         reflect.Type
	IndirectFieldType reflect.Type
	StructField       reflect.StructField
	Tag               *Tag
	Table             *Table
}
