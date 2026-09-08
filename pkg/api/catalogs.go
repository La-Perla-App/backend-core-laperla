package api

import (
	"connectrpc.com/connect"
)

var (
	generalParamsHeaderKey = "General-Params-Bin"
	ResponseInfoHeaderKey  = "Response-Info-Bin"
	ClientIDCookieName     = "laperla-client-id"
)

var ResponseTypes = struct {
	Error   ResponseType
	Success ResponseType
}{
	Error:   "error",
	Success: "success",
}

type APIErrorCode uint

const (
	InternalServerErrorCode   = APIErrorCode(connect.CodeInternal)
	UnauthorizedCode          = APIErrorCode(connect.CodeUnauthenticated)
	NotFoundCode              = APIErrorCode(connect.CodeNotFound)
	InvalidRequestDataCode    = APIErrorCode(connect.CodeInvalidArgument)
	AlreadyExistsCode         = APIErrorCode(connect.CodeAlreadyExists)
	TokenInvalidOrExpiredCode = APIErrorCode(427)
	UserIsDeactivatedCode     = APIErrorCode(428)
	UserNotVerifiedCode       = APIErrorCode(429)
)

var APIErrorCodesMessages = map[APIErrorCode]string{
	InternalServerErrorCode:   ErrorMessages.InternalServerError,
	UnauthorizedCode:          ErrorMessages.Unauthorized,
	NotFoundCode:              ErrorMessages.NotFound,
	InvalidRequestDataCode:    ErrorMessages.InvalidRequestData,
	AlreadyExistsCode:         ErrorMessages.AlreadyExists,
	TokenInvalidOrExpiredCode: ErrorMessages.TokenInvalidOrExpired,
	UserIsDeactivatedCode:     ErrorMessages.UserIsDeactivated,
	UserNotVerifiedCode:       ErrorMessages.UserNotVerified,
}

func GetErrorMessageFromCode(code APIErrorCode) string {
	if msg, ok := APIErrorCodesMessages[code]; ok {
		return msg
	}
	return ErrorMessages.InternalServerError
}

var ErrorMessages = struct {
	InternalServerError   string
	Unauthorized          string
	NotFound              string
	InvalidRequestData    string
	AlreadyExists         string
	TokenInvalidOrExpired string
	UserIsDeactivated     string
	UserNotVerified       string
}{
	InternalServerError:   "internal_server_error",
	Unauthorized:          "invalid_token",
	NotFound:              "not_found",
	InvalidRequestData:    "invalid_request_data",
	AlreadyExists:         "already_exists",
	TokenInvalidOrExpired: "token_invalid_or_expired",
	UserIsDeactivated:     "user_is_deactivated",
	UserNotVerified:       "user_not_verified",
}

// DefaultSuccessMessage default success message.
var DefaultSuccessMessage = "Success"

const (
	PlatformAndroid = "android"
	PlatformIOS     = "ios"
	PlatformWeb     = "web"
	PlatformUnknown = "unknown"
)
