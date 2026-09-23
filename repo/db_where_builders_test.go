package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Grant filters (type, user_id) come from raw HTTP query params; they must be
// bound, never interpolated, and the order direction can't be injected.
func Test_buildAndGetWhereGrantQuery_Parameterized(t *testing.T) {
	inj := "x' OR '1'='1"
	cancelled := true
	where, orderBy, args := buildAndGetWhereGrantQuery(&cancelled, inj, "hhmembership", "asc")

	assert.NotContains(t, where, inj, "raw input must not be interpolated into the query")
	assert.Contains(t, where, "cancelled_at IS NOT NULL") // typed bool → fixed literal
	assert.Contains(t, where, "type=$1")
	assert.Contains(t, where, "user_id=$2")
	assert.Equal(t, []interface{}{"hhmembership", inj}, args)
	assert.Equal(t, " ORDER BY created_at asc", orderBy)

	// Order direction can't be injected: anything but desc collapses to asc.
	_, ob, _ := buildAndGetWhereGrantQuery(nil, "", "", "created_at; DROP TABLE users")
	assert.Equal(t, " ORDER BY created_at asc", ob)

	// No filters → no WHERE, no args, default order.
	w, ob2, a := buildAndGetWhereGrantQuery(nil, "", "", "")
	assert.Empty(t, w)
	assert.Empty(t, a)
	assert.Equal(t, " ORDER BY updated_at desc", ob2)
}

// Membership user_id comes from a raw HTTP query param; it must be bound.
func Test_buildAndGetWhereMembershipQuery_Parameterized(t *testing.T) {
	inj := "x' OR '1'='1"
	where, orderBy, args := buildAndGetWhereMembershipQuery(inj)

	assert.NotContains(t, where, inj, "raw input must not be interpolated into the query")
	assert.Contains(t, where, "user_id=$1")
	assert.Equal(t, []interface{}{inj}, args)
	assert.Equal(t, " ORDER BY updated_at desc", orderBy)

	// No filter → no WHERE, no args.
	w, _, a := buildAndGetWhereMembershipQuery("")
	assert.Empty(t, w)
	assert.Empty(t, a)
}
