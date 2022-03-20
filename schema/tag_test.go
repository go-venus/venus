package schema

import (
	"fmt"
	"reflect"
	"testing"
)

type TagTest struct {
	Id   string `lighthouse:"id;PrimaryKey;SIZE:2"`
	Name string `lighthouse:"name"`
}

func TestParseTag(t *testing.T) {
	typ := reflect.Indirect(reflect.ValueOf(&TagTest{})).Type()
	num := typ.NumField()
	for i := 0; i < num; i++ {
		field := typ.Field(i)

		tag := ParseTag(field.Tag, ";")
		fmt.Println(tag)
	}
}
