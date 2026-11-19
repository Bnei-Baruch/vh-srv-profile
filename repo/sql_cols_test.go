package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_sqlCols(t *testing.T) {
	var c sqlCols
	assert.True(t, c.empty())

	name := "o'name"
	var missing *int
	c.set("a", 1)
	setIf(&c, "name", &name)
	setIf(&c, "missing", missing)
	assert.False(t, c.empty())

	assert.Equal(t, "a=$1,name=$2", c.assignments())
	assert.Equal(t, "(a,name) VALUES ($1,$2)", c.values())
	assert.Equal(t, "$3", c.bind(42))
	assert.Equal(t, []any{1, "o'name", 42}, c.args)
}
