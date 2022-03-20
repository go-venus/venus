package clause

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSelect(t *testing.T) {
	var clause Clause
	clause.Set(Limit, 3)
	clause.Set(Select, "User", []string{"*"})
	clause.Set(Where, "StructName = ?", "Tom")
	clause.Set(OrderBy, "Age ASC")
	sql, vars := clause.Build(Select, Where, OrderBy, Limit)
	t.Log(sql, vars)
	if sql != "SELECT * FROM User WHERE StructName = ? ORDER BY Age ASC LIMIT ?" {
		t.Fatal("failed to build SQL")
	}
	if !reflect.DeepEqual(vars, []interface{}{"Tom", 3}) {
		t.Fatal("failed to build SQLVars")
	}
}
func TestCount(t *testing.T) {
	var clause Clause
	clause.Set(Count, "User")
	clause.Set(Select, "User", []string{"*"})
	sql, vars := clause.Build(Select, Select)
	fmt.Println(sql, vars)

}
