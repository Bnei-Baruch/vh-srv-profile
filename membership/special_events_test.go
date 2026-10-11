package membership

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
)

func strOrNil(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

// Orders stores a special's absent email or keycloak_id as NULL, and its
// events carry that as null, "" or no field at all. The old code asserted
// .(string) unchecked: a null panicked the consumer and the acked event was
// lost.
func TestGetUserIDs_SpecialEvents(t *testing.T) {
	tests := []struct {
		name      string
		event     string
		wantEmail any
		wantKC    any
		wantErr   bool
	}{
		{
			name:   "delete, keycloak-only special",
			event:  `{"type":"delete_special","payload":{"email":null,"keycloak_id":"kc-1"}}`,
			wantKC: "kc-1",
		},
		{
			name:   "delete, keycloak-only special, email sent as empty",
			event:  `{"type":"delete_special","payload":{"email":"","keycloak_id":"kc-1"}}`,
			wantKC: "kc-1",
		},
		{
			name:      "delete from an orders without keycloak_id in the payload",
			event:     `{"type":"delete_special","payload":{"email":"a@test.test"}}`,
			wantEmail: "a@test.test",
		},
		{
			name:      "delete, both identifiers",
			event:     `{"type":"delete_special","payload":{"email":"a@test.test","keycloak_id":"kc-1"}}`,
			wantEmail: "a@test.test",
			wantKC:    "kc-1",
		},
		{
			name:    "delete, no identifier",
			event:   `{"type":"delete_special","payload":{"email":null}}`,
			wantErr: true,
		},
		{
			name:   "create, null email",
			event:  `{"type":"create_special","payload":{"email":null,"keycloak_id":"kc-1"}}`,
			wantKC: "kc-1",
		},
		{
			name:    "create, no identifier",
			event:   `{"type":"create_special","payload":{"email":"","keycloak_id":null}}`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var event orders.Event
			require.NoError(t, json.Unmarshal([]byte(tc.event), &event))

			eh := &EventsHandler{}
			ids, err := eh.getUserIDs(context.Background(), event)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantEmail, strOrNil(ids.Email))
			assert.Equal(t, tc.wantKC, strOrNil(ids.KeycloakID))
			assert.Nil(t, ids.UserID)
		})
	}
}

// Every id in an orders event is read with a checked assertion: a missing or
// null id is an error for that event, not a panic that loses it.
func TestOrdersEvents_MissingIDsAreErrors(t *testing.T) {
	for _, raw := range []string{
		`{"type":"create_order","payload":{}}`,
		`{"type":"update_order","payload":{"order_id":null}}`,
		`{"type":"create_payment","payload":{}}`,
		`{"type":"delete_payment","payload":{"payment_id":null}}`,
	} {
		var event orders.Event
		require.NoError(t, json.Unmarshal([]byte(raw), &event))
		eh := &EventsHandler{}
		_, err := eh.getUserIDs(context.Background(), event)
		assert.Error(t, err, raw)
	}

	for _, raw := range []string{
		`{"type":"merge_account","payload":{}}`,
		`{"type":"merge_account","payload":{"source_account_id":"1","destination_account_id":null}}`,
		`{"type":"merge_account","payload":{"source_account_id":null,"destination_account_id":"2"}}`,
	} {
		var event orders.Event
		require.NoError(t, json.Unmarshal([]byte(raw), &event))
		// No repo: reaching MergeAccounts would panic on the nil interface.
		eh := &EventsHandler{}
		_, err := eh.mergeAccounts(context.Background(), event)
		assert.Error(t, err, raw)
	}
}

// A special event with no user is expected until vh-srv-orders#39 is deployed
// (a special with no email). It is skipped with a warning, not reported to
// Sentry as an error nobody can act on.
func TestGetUserIDs_SpecialWithoutUserIsSkipped(t *testing.T) {
	var event orders.Event
	require.NoError(t, json.Unmarshal([]byte(`{"type":"delete_special","payload":{"email":null}}`), &event))

	_, err := (&EventsHandler{}).getUserIDs(context.Background(), event)
	assert.ErrorIs(t, err, errSkip)
}
