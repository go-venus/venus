package schema

import (
	"testing"
	"time"
)

type Model struct {
	Id   int
	time time.Time
}
type User struct {
	Model
	Name string `venus:"PRIMARY KEY"`
	Age  int
}

func TestParse(t *testing.T) {
	var TestDial, _ = GetDialect("mysql")

	schema := Parse(&User{}, TestDial)
	if schema.TableName != "User" || len(schema.Fields) != 2 {
		t.Fatal("failed to parse User struct")
	}
	if schema.GetField("StructName").Tag.Tag != `venus:"PRIMARY KEY"` {
		t.Fatal("failed to parse primary key")
	}
}
