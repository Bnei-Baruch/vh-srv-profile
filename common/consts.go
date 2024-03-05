package common

const (
	ServiceName = "vh-srv-profile"

	CtxTokenSource = "TOKEN_SOURCE"

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
)
