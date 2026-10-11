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
