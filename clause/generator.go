package clause

import (
	"fmt"
	"strings"
)

type generator func(values ...interface{}) (string, []interface{})

var generators = make(map[Type]generator)

func init() {
	generators[Insert] = generatorInsert
	generators[Values] = generatorValues
	generators[Select] = generatorSelect
	generators[Limit] = generatorLimit
	generators[Where] = generatorWhere
	generators[OrderBy] = generatorOrderBy
	generators[Update] = generatorUpdate
	generators[Delete] = generatorDelete
	generators[Count] = generatorCount
}

func generatorCount(values ...interface{}) (string, []interface{}) {
	return generatorSelect(values[0], []string{"count(*)"})
}

func generatorDelete(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

func generatorUpdate(values ...interface{}) (string, []interface{}) {
	// UPDATE $tableName SET ($fields)
	tableName := values[0]
	param := values[1].(map[string]interface{})
	var keys []string
	var vars []interface{}
	for k, v := range param {
		keys = append(keys, k+" = ?")
		vars = append(vars, v)
	}
	return fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(keys, ", ")), vars

}

func generatorInsert(values ...interface{}) (string, []interface{}) {
	// INSERT INTO $tableName ($fields)
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s (%s)", tableName, fields), []interface{}{}
}

func generatorValues(values ...interface{}) (string, []interface{}) {
	// VALUES ($v1), ($v2), ...
	var bindStr string
	var sql strings.Builder
	var vars []interface{}
	sql.WriteString("VALUES ")
	for i, value := range values {
		v := value.([]interface{})
		if bindStr == "" {
			bindStr = genBindVars(len(v))
		}
		sql.WriteString(fmt.Sprintf("(%s)", bindStr))
		if i+1 != len(values) {
			sql.WriteString(", ")
		}
		vars = append(vars, v...)
	}
	return sql.String(), vars
}

func generatorSelect(values ...interface{}) (string, []interface{}) {
	// SELECT $fields FROM $tableName
	tableName := values[0].(string)
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %s FROM %s", fields, tableName), []interface{}{}
}

func generatorLimit(values ...interface{}) (string, []interface{}) {
	// LIMIT $num
	return "LIMIT ?", values
}

func generatorWhere(values ...interface{}) (string, []interface{}) {
	// WHERE $desc
	desc, vars := values[0], values[1:]
	return fmt.Sprintf("WHERE %s", desc), vars
}

func generatorOrderBy(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}

func genBindVars(n int) string {
	switch n {
	case 0:
		return ""
	case 1:
		return "?"
	}

	builder := strings.Builder{}
	builder.Grow(2*n + 2*(n-1))
	builder.WriteRune('?')
	for i := 0; i < n-1; i++ {
		builder.WriteString(", ")
		builder.WriteRune('?')
	}
	return builder.String()
}
