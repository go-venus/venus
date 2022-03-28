package dialect

import (
	"errors"
	"reflect"
	"sync"
)

var (
	rw          = sync.RWMutex{}
	dialectsMap = map[string]Dialect{}
)

var ErrNotFoundDialect = errors.New("not found dialect")

// Dialect getDB Dialect
type Dialect interface {
	DataTypeOf(v reflect.Value) string
	TableExistSQL(tableName string) (string, []any)
}

// RegisterDialect Register Dialect.
func RegisterDialect(name string, dialect Dialect) {
	rw.RLock()
	defer rw.RUnlock()

	dialectsMap[name] = dialect
}

// GetDialect Get Dialect.
func GetDialect(name string) (dialect Dialect, err error) {
	rw.Lock()
	defer rw.Unlock()

	var ok bool
	if dialect, ok = dialectsMap[name]; !ok {
		err = ErrNotFoundDialect
	}

	return
}
