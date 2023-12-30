package yubico

const (
	StatusOK                  = "OK"
	StatusBadOTP              = "BAD_OTP"
	StatusReplayedOTP         = "REPLAYED_OTP"
	StatusBadSignature        = "BAD_SIGNATURE"
	StatusMissingParam        = "MISSING_PARAMETER"
	StatusNoSuchClient        = "NO_SUCH_CLIENT"
	StatusOperationNotAllowed = "OPERATION_NOT_ALLOWED"
	StatusBackendError        = "BACKEND_ERROR"
	StatusNotEnoughAnswers    = "NOT_ENOUGH_ANSWERS"
	StatusReplayedRequest     = "REPLAYED_REQUEST"
)
