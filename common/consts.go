package common

const (
	ServiceName = "vh-srv-profile"

	CtxTokenSource = "TOKEN_SOURCE"
	CtxRequestID   = "REQUEST_ID"
	CtxLogger      = "LOGGER"
	CtxAuthClaims  = "AUTH_CLAIMS"

	RoleRoot           = "vh_root" // kong service clients has this role as well to allow inter-service communication
	RoleAdmin          = "vh_admin"
	RoleHelpHaverAdmin = "vh_helphaver_admin"

	RequestTypeHelpHaver   = "hhmembership"
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

	MembershipGracePeriodInDays = 7
)

var RoleAnyAdmin = []string{RoleRoot, RoleAdmin, RoleHelpHaverAdmin}

// This gets set at build time via `-ldflags "-X ..."`
var GitSHA string = "local"
