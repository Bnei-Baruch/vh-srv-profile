package events

import (
	"io"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

const (
	ActorSystem = common.ServiceName

	ComponentAPI                          = "api"
	ComponentMembershipInvalidator        = "membership_invalidator"
	ComponentMembershipMigrator           = "membership_migrator"
	ComponentMembershipBulkEval           = "membership_bulk_eval"
	ComponentMembershipOrdersEventHandler = "membership_orders_event_handler"
	ComponentUserSyncer                   = "user_syncer"

	TypeCreateProfile     = "create_profile"
	TypeUpdateProfile     = "update_profile"
	TypeDeleteProfile     = "delete_profile"
	TypeHardDeleteProfile = "hard_delete_profile"
	TypeMergeAccounts     = "merge_accounts"
)

var entropy io.Reader

func init() {
	entropy = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
}

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Component string                 `json:"component"`
	Actor     string                 `json:"actor"`
	RequestID string                 `json:"request_id,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
}

type EventBuilder interface {
	BuildEvent(eventType string, payload map[string]interface{}) Event
}

func MakeEvent(eventType string, payload map[string]interface{}) Event {
	return Event{
		ID:        ulid.MustNew(ulid.Now(), entropy).String(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}
