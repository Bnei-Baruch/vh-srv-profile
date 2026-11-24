package repo

import (
	"strconv"
	"strings"
)

// sqlCols collects column/value pairs for an INSERT or UPDATE that writes
// only the fields a caller set. Column names are always literals at the
// call site; every value is bound as $n.
type sqlCols struct {
	cols []string
	args []any
}

// set adds col with value v.
func (c *sqlCols) set(col string, v any) {
	c.cols = append(c.cols, col)
	c.args = append(c.args, v)
}

// setIf adds col with *v when v is not nil.
func setIf[T any](c *sqlCols, col string, v *T) {
	if v != nil {
		c.set(col, *v)
	}
}

func (c *sqlCols) empty() bool { return len(c.cols) == 0 }

// assignments returns "a=$1,b=$2" for UPDATE ... SET.
func (c *sqlCols) assignments() string {
	parts := make([]string, len(c.cols))
	for i, col := range c.cols {
		parts[i] = col + "=$" + strconv.Itoa(i+1)
	}
	return strings.Join(parts, ",")
}

// values returns "(a,b) VALUES ($1,$2)" for INSERT INTO t ...
func (c *sqlCols) values() string {
	ph := make([]string, len(c.cols))
	for i := range c.cols {
		ph[i] = "$" + strconv.Itoa(i+1)
	}
	return "(" + strings.Join(c.cols, ",") + ") VALUES (" + strings.Join(ph, ",") + ")"
}

// bind adds one more value after the columns (e.g. the WHERE id) and returns
// its placeholder. Call it after the last set.
func (c *sqlCols) bind(v any) string {
	c.args = append(c.args, v)
	return "$" + strconv.Itoa(len(c.args))
}
