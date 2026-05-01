package common

const (
	ServiceName = "vh-srv-profile"

	CtxEventBuilder = "EVENT_BUILDER"
	CtxRequestID    = "REQUEST_ID"
	CtxLogger       = "LOGGER"
	CtxTokenSource  = "TOKEN_SOURCE"
	CtxAuthClaims   = "AUTH_CLAIMS"

	RoleRoot           = "vh_root" // kong service clients has this role as well to allow inter-service communication
	RoleAdmin          = "vh_admin"
	RoleHelpHaverAdmin = "vh_helphaver_admin"

	RequestTypeHelpHaverLegacy = "hhmembership"
	RequestTypeHHGimlaj  = "hh-gimlaj"
	RequestTypeHHHayal   = "hh-hayal"
	RequestTypeHHOther   = "hh-other"
	RequestStatusRequested = "REQUESTED"
	RequestStatusApproved  = "APPROVED"
	RequestStatusDenied    = "DENIED"

	GrantTypeMembershipMonths = "mb_months"

	NotificationSlugMBNew                    = "mb_new"
	NotificationSlugMBCancelled              = "mb_cancelled"
	NotificationSlugMBExpirationNotice       = "mb_expiration_notice"
	NotificationSlugMBHasExpiredNotice       = "mb_has_expired_notice"
	NotificationSlugMBProblemPreviousPayment = "mb_problem_previous_payment"
	NotificationSlugHHRequestReceived        = "hh_request_received"
	NotificationSlugHHRequestApproved        = "hh_request_approved"
	NotificationSlugHHRequestRefused         = "hh_request_refused"

	Mobile   = "mobile"
	WhatsApp = "WhatsApp"
	Telegram = "Telegram"

	MaritalStatusSingle   = "Single"
	MaritalStatusMarried  = "Married"
	MaritalStatusDivorced = "Divorced"

	MembershipGracePeriodInDays = 7
)

var RoleAnyAdmin = []string{RoleRoot, RoleAdmin, RoleHelpHaverAdmin}

// All HH membership request types (legacy + new)
var RequestTypeHHAll = []string{RequestTypeHelpHaverLegacy, RequestTypeHHGimlaj, RequestTypeHHHayal, RequestTypeHHOther}

func IsValidHHType(t string) bool {
	for _, v := range RequestTypeHHAll {
		if t == v {
			return true
		}
	}
	return false
}

// This gets set at build time via `-ldflags "-X ..."`
var GitSHA string = "local"
