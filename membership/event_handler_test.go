package membership

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

// notifyMock is a focused ProfileRepository mock that records NotifyHHRequest
// calls. It embeds the interface so it satisfies the full type while only
// implementing the single method handleHHRequestEvent uses (the project-wide
// storageMock lives in the api package and isn't importable here).
type notifyMock struct {
	repo.ProfileRepository
	calls   int
	gotKC   string
	gotSlug string
	err     error
}

func (m *notifyMock) NotifyHHRequest(_ context.Context, keycloakID, slug string) error {
	m.calls++
	m.gotKC = keycloakID
	m.gotSlug = slug
	return m.err
}

func TestHandleHHRequestEvent_SlugSelection(t *testing.T) {
	const kc = "5e47d90f-b00e-44cd-b7c0-0c32d2521da7"

	tests := []struct {
		name     string
		event    orders.Event
		wantSlug string
	}{
		{
			name:     "created -> received",
			event:    orders.Event{Type: orders.TypeHHRequestCreated, Payload: map[string]interface{}{"keycloak_id": kc}},
			wantSlug: common.NotificationSlugHHRequestReceived,
		},
		{
			name:     "concluded approved -> approved",
			event:    orders.Event{Type: orders.TypeHHRequestConcluded, Payload: map[string]interface{}{"keycloak_id": kc, "approved": true}},
			wantSlug: common.NotificationSlugHHRequestApproved,
		},
		{
			name:     "concluded not approved -> refused",
			event:    orders.Event{Type: orders.TypeHHRequestConcluded, Payload: map[string]interface{}{"keycloak_id": kc, "approved": false}},
			wantSlug: common.NotificationSlugHHRequestRefused,
		},
		{
			name:     "concluded missing approved -> refused (safe default)",
			event:    orders.Event{Type: orders.TypeHHRequestConcluded, Payload: map[string]interface{}{"keycloak_id": kc}},
			wantSlug: common.NotificationSlugHHRequestRefused,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &notifyMock{}
			eh := &EventsHandler{repo: m}

			eh.handleHHRequestEvent(context.Background(), tc.event)

			assert.Equal(t, 1, m.calls, "NotifyHHRequest should be called exactly once")
			assert.Equal(t, kc, m.gotKC)
			assert.Equal(t, tc.wantSlug, m.gotSlug)
		})
	}
}

func TestHandleHHRequestEvent_MissingKeycloakID(t *testing.T) {
	m := &notifyMock{}
	eh := &EventsHandler{repo: m}

	// No keycloak_id in the payload -> the guard must skip without notifying.
	eh.handleHHRequestEvent(context.Background(), orders.Event{
		Type:    orders.TypeHHRequestCreated,
		Payload: map[string]interface{}{},
	})

	assert.Equal(t, 0, m.calls, "NotifyHHRequest must not be called without a keycloak_id")
}
