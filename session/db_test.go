package session

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-venus/venus/schema"
)

func TestNew(t *testing.T) {
	type User struct {
		Name string `venus:"name"`
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// mock.ExpectBegin()
	// mock.ExpectExec("UPDATE products").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `user` (name) VALUES (?)").WithArgs("1").WillReturnResult(sqlmock.NewResult(1, 1))
	// mock.ExpectCommit()

	var TestDial, _ = schema.GetDialect("mysql")

	s := New[User](db, TestDial)
	s.Insert(&User{Name: "1"})
}
