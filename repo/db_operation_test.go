package repo

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

func insertOperationUser(t *testing.T, db *ProfileDB, userID, kc, email string) {
	t.Helper()
	_, err := db.Exec(testContext(), `
	INSERT INTO users (user_id, keycloak_id, first_name_vernacular, last_name_vernacular, primary_email)
	VALUES ($1, $2, 'first', 'last', $3)`, userID, kc, email)
	require.NoError(t, err)
}

func userEmailAndKc(t *testing.T, db *ProfileDB, userID string) (string, string) {
	t.Helper()
	var email, kc string
	require.NoError(t, db.QueryRow(testContext(),
		`SELECT primary_email, keycloak_id FROM users WHERE user_id=$1`, userID).Scan(&email, &kc))
	return email, kc
}

// Emails with an apostrophe go through the rename and its revert unchanged,
// and the revert row records {sql,args}.
func Test_ProfileDB_PerformOperation_RevertOperation_email_with_apostrophe(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	const (
		userID   = "22000000-0000-0000-0000-000000000001"
		kc       = "11000000-0000-0000-0000-000000000001"
		oldEmail = "o'old@test.test"
		newEmail = "o'new@test.test"
	)
	insertOperationUser(t, db, userID, kc, oldEmail)

	id, err := db.PerformOperation(ctx, OperationReq{
		NewEmail:      utils.PointerString(newEmail),
		OldEmail:      utils.PointerString(oldEmail),
		NewKeycloakID: utils.PointerString(kc),
		Type:          utils.PointerString("email_update"),
	})
	require.NoError(t, err)
	assert.NotZero(t, id)

	email, gotKc := userEmailAndKc(t, db, userID)
	assert.Equal(t, newEmail, email)
	assert.Equal(t, kc, gotKc)

	var status string
	var revert QueryLog
	require.NoError(t, db.QueryRow(ctx, `SELECT status, revert FROM operation_trace WHERE id=$1`, id).
		Scan(&status, &revert))
	assert.Equal(t, "success", status)
	require.Len(t, revert.Queries, 1)
	assert.Equal(t, []string{oldEmail, newEmail, kc}, revert.Queries[0].Args)
	assert.NotContains(t, revert.Queries[0].SQL, "test.test", "values must be args, not SQL text")

	require.NoError(t, db.RevertOperation(ctx, newEmail, oldEmail))

	email, _ = userEmailAndKc(t, db, userID)
	assert.Equal(t, oldEmail, email)

	require.NoError(t, db.QueryRow(ctx, `SELECT status FROM operation_trace WHERE id=$1`, id).Scan(&status))
	assert.Equal(t, "reverted", status)

	require.ErrorIs(t, db.RevertOperation(ctx, newEmail, oldEmail), common.ErrOperationReverted)
}

// With an old keycloak id, both keycloak_id and email move, and revert moves them back.
func Test_ProfileDB_PerformOperation_RevertOperation_with_old_keycloak_id(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	const (
		userID   = "22000000-0000-0000-0000-000000000002"
		oldKc    = "11000000-0000-0000-0000-000000000002"
		newKc    = "11000000-0000-0000-0000-000000000003"
		oldEmail = "o'old@test.test"
		newEmail = "o'new@test.test"
	)
	insertOperationUser(t, db, userID, oldKc, oldEmail)

	_, err := db.PerformOperation(ctx, OperationReq{
		NewEmail:      utils.PointerString(newEmail),
		OldEmail:      utils.PointerString(oldEmail),
		NewKeycloakID: utils.PointerString(newKc),
		OldKeycloakID: utils.PointerString(oldKc),
		Type:          utils.PointerString("email_update"),
	})
	require.NoError(t, err)

	email, kc := userEmailAndKc(t, db, userID)
	assert.Equal(t, newEmail, email)
	assert.Equal(t, newKc, kc)

	require.NoError(t, db.RevertOperation(ctx, newEmail, oldEmail))

	email, kc = userEmailAndKc(t, db, userID)
	assert.Equal(t, oldEmail, email)
	assert.Equal(t, oldKc, kc)
}

// Reverts racing on the same operation: the row lock lets exactly one through.
func Test_ProfileDB_RevertOperation_concurrent_only_one_succeeds(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	const (
		userID   = "22000000-0000-0000-0000-000000000004"
		kc       = "11000000-0000-0000-0000-000000000004"
		oldEmail = "old@test.test"
		newEmail = "new@test.test"
	)
	insertOperationUser(t, db, userID, kc, oldEmail)

	_, err := db.PerformOperation(ctx, OperationReq{
		NewEmail:      utils.PointerString(newEmail),
		OldEmail:      utils.PointerString(oldEmail),
		NewKeycloakID: utils.PointerString(kc),
	})
	require.NoError(t, err)

	const n = 20
	errs := make([]error, n)
	var start, done sync.WaitGroup
	start.Add(1)
	for i := 0; i < n; i++ {
		done.Add(1)
		go func(i int) {
			defer done.Done()
			start.Wait()
			errs[i] = db.RevertOperation(ctx, newEmail, oldEmail)
		}(i)
	}
	start.Done()
	done.Wait()

	succeeded := 0
	for _, e := range errs {
		if e == nil {
			succeeded++
		} else {
			assert.Contains(t, e.Error(), "already reverted")
		}
	}
	assert.Equal(t, 1, succeeded)

	email, _ := userEmailAndKc(t, db, userID)
	assert.Equal(t, oldEmail, email)
}

func Test_ProfileDB_RevertOperation_not_found(t *testing.T) {
	db := newTestProfileDBIsolated(t)

	err := db.RevertOperation(testContext(), "nobody@test.test", "o'nobody@test.test")
	assert.ErrorIs(t, err, common.ErrOperationNotFound)
}

// The revert builds its statement from the stored input; SQL in the revert
// column is a record only and never runs.
func Test_ProfileDB_RevertOperation_does_not_run_stored_sql(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	const (
		userID = "22000000-0000-0000-0000-000000000009"
		kc     = "11000000-0000-0000-0000-000000000009"
	)
	insertOperationUser(t, db, userID, kc, "old@test.test")

	_, err := db.PerformOperation(ctx, OperationReq{
		NewEmail:      utils.PointerString("new@test.test"),
		OldEmail:      utils.PointerString("old@test.test"),
		NewKeycloakID: utils.PointerString(kc),
		Type:          utils.PointerString("email_update"),
	})
	require.NoError(t, err)

	_, err = db.Exec(ctx, `UPDATE operation_trace SET revert =
		'{"queries":[{"sql":"UPDATE users SET first_name_vernacular = ''pwned''","args":[]}],"logs":[]}'`)
	require.NoError(t, err)

	require.NoError(t, db.RevertOperation(ctx, "new@test.test", "old@test.test"))

	email, _ := userEmailAndKc(t, db, userID)
	assert.Equal(t, "old@test.test", email)
	var pwned int
	require.NoError(t, db.QueryRow(ctx, `SELECT count(*) FROM users WHERE first_name_vernacular = 'pwned'`).Scan(&pwned))
	assert.Zero(t, pwned)
}

// A rename whose old email and key match no user changes nothing and writes
// no trace, instead of reporting success for UPDATE 0.
func Test_ProfileDB_PerformOperation_no_matching_user_changes_nothing(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	const kc = "11000000-0000-0000-0000-000000000010"
	insertOperationUser(t, db, "22000000-0000-0000-0000-000000000010", kc, "old@test.test")

	_, err := db.PerformOperation(ctx, OperationReq{
		NewEmail:      utils.PointerString("new@test.test"),
		OldEmail:      utils.PointerString("typo@test.test"),
		NewKeycloakID: utils.PointerString(kc),
		Type:          utils.PointerString("email_update"),
	})
	require.ErrorIs(t, err, common.ErrOperationNoMatch)

	var traces int
	require.NoError(t, db.QueryRow(ctx, `SELECT count(*) FROM operation_trace`).Scan(&traces))
	assert.Zero(t, traces)
}

// If the user moved off the new key after the rename, the revert matches no
// user; it is refused and the trace stays "success".
func Test_ProfileDB_RevertOperation_user_changed_since_changes_nothing(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	const (
		userID = "22000000-0000-0000-0000-000000000011"
		oldKc  = "11000000-0000-0000-0000-000000000011"
		newKc  = "11000000-0000-0000-0000-000000000012"
		other  = "11000000-0000-0000-0000-000000000013"
	)
	insertOperationUser(t, db, userID, oldKc, "a@test.test")

	_, err := db.PerformOperation(ctx, OperationReq{
		NewEmail:      utils.PointerString("b@test.test"),
		OldEmail:      utils.PointerString("a@test.test"),
		OldKeycloakID: utils.PointerString(oldKc),
		NewKeycloakID: utils.PointerString(newKc),
		Type:          utils.PointerString("email_update"),
	})
	require.NoError(t, err)
	_, err = db.Exec(ctx, `UPDATE users SET keycloak_id = $1 WHERE user_id = $2`, other, userID)
	require.NoError(t, err)

	require.ErrorIs(t, db.RevertOperation(ctx, "b@test.test", "a@test.test"), common.ErrOperationNoMatch)

	email, kc := userEmailAndKc(t, db, userID)
	assert.Equal(t, []string{"b@test.test", other}, []string{email, kc})
	var status string
	require.NoError(t, db.QueryRow(ctx, `SELECT status FROM operation_trace`).Scan(&status))
	assert.Equal(t, "success", status)
}

// Moving to a keycloak id another user has is a conflict, not a database
// failure, in both directions; nothing changes.
func Test_ProfileDB_operation_taken_keycloak_id_is_conflict(t *testing.T) {
	db := newTestProfileDBIsolated(t)
	ctx := testContext()

	const (
		userID = "22000000-0000-0000-0000-000000000021"
		oldKc  = "11000000-0000-0000-0000-000000000021"
		newKc  = "11000000-0000-0000-0000-000000000022"
	)
	insertOperationUser(t, db, userID, oldKc, "a@test.test")
	insertOperationUser(t, db, "22000000-0000-0000-0000-000000000022", newKc, "taken@test.test")

	rename := OperationReq{
		NewEmail:      utils.PointerString("b@test.test"),
		OldEmail:      utils.PointerString("a@test.test"),
		OldKeycloakID: utils.PointerString(oldKc),
		NewKeycloakID: utils.PointerString(newKc),
		Type:          utils.PointerString("email_update"),
	}
	_, err := db.PerformOperation(ctx, rename)
	require.ErrorIs(t, err, common.ErrOperationConflict)

	// Free the new key, rename, then give the old key to someone else.
	_, err = db.Exec(ctx, `DELETE FROM users WHERE keycloak_id = $1`, newKc)
	require.NoError(t, err)
	_, err = db.PerformOperation(ctx, rename)
	require.NoError(t, err)
	insertOperationUser(t, db, "22000000-0000-0000-0000-000000000023", oldKc, "again@test.test")

	require.ErrorIs(t, db.RevertOperation(ctx, "b@test.test", "a@test.test"), common.ErrOperationConflict)
	email, kc := userEmailAndKc(t, db, userID)
	assert.Equal(t, []string{"b@test.test", newKc}, []string{email, kc})
}
