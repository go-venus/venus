package clause

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_genBindVars(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		bindVars := genBindVars(1)
		assert.EqualValues(t, "?", bindVars)

	})

	t.Run("2", func(t *testing.T) {
		bindVars := genBindVars(2)
		assert.EqualValues(t, "?, ?", bindVars)
	})

}

func FuzzGenbindvars(f *testing.F) {
	f.Add(0)
	f.Fuzz(func(t *testing.T, n int) {
		if n < 0 {
			t.Skip()
		}
		vars := genBindVars(n)
		var size int
		if n != 0 {
			size = n + 2*(n-1)
		}
		assert.Len(t, vars, size)
	})
}
