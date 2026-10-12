package api

import (
	"errors"
	"net/http"

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
	InternalServerErrorCode = APIErrorCode(connect.CodeInternal)
	UnauthorizedCode        = APIErrorCode(connect.CodeUnauthenticated)
	NotFoundCode            = APIErrorCode(connect.CodeNotFound)
	InvalidRequestDataCode  = APIErrorCode(connect.CodeInvalidArgument)
	AlreadyExistsCode       = APIErrorCode(connect.CodeAlreadyExists)
	// 427 es el token de un enlace (verificar correo, recuperar contraseña),
	// no el de la sesión.
	TokenInvalidOrExpiredCode = APIErrorCode(427)
	UserIsDeactivatedCode     = APIErrorCode(428)
	UserNotVerifiedCode       = APIErrorCode(429)
	// La sesión ya no vive: el cliente tiene que volver a entrar. Distinto de
	// unas credenciales mal escritas, que no implican cerrar nada.
	SessionExpiredCode = APIErrorCode(430)
	// Correo o contraseña incorrectos. Deliberadamente el MISMO código para
	// los dos: distinguirlos le diría a cualquiera qué correos tienen cuenta.
	InvalidCredentialsCode = APIErrorCode(431)
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
	SessionExpiredCode:        ErrorMessages.SessionExpired,
	InvalidCredentialsCode:    ErrorMessages.InvalidCredentials,
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
	SessionExpired        string
	InvalidCredentials    string
}{
	InternalServerError:   "internal_server_error",
	Unauthorized:          "invalid_token",
	NotFound:              "not_found",
	InvalidRequestData:    "invalid_request_data",
	AlreadyExists:         "already_exists",
	TokenInvalidOrExpired: "token_invalid_or_expired",
	UserIsDeactivated:     "user_is_deactivated",
	UserNotVerified:       "user_not_verified",
	SessionExpired:        "session_expired",
	InvalidCredentials:    "invalid_credentials",
}

// DefaultSuccessMessage default success message.
var DefaultSuccessMessage = "Success"

const (
	PlatformAndroid = "android"
	PlatformIOS     = "ios"
	PlatformWeb     = "web"
	PlatformUnknown = "unknown"
)

// ErrorMessageFor es el mensaje de un error que nadie nombró antes: el del
// catálogo para su código (`invalid_request_data`, `not_found`…) y, si el
// catálogo no lo tiene, el nombre del código Connect (`permission_denied`,
// `failed_precondition`…). Sólo un fallo de verdad queda en
// `internal_server_error`.
//
// Antes cualquier error sin mensaje explícito salía como
// `internal_server_error`, también un 400 por falta de un parámetro: el
// envoltorio REST decía «se rompió algo» cuando lo que había era una
// petición mal hecha.
func ErrorMessageFor(err error) string {
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		return ErrorMessages.InternalServerError
	}
	code := connectErr.Code()
	if msg, ok := APIErrorCodesMessages[APIErrorCode(code)]; ok {
		return msg
	}
	switch code {
	case connect.CodeUnknown, connect.CodeDataLoss:
		return ErrorMessages.InternalServerError
	}
	return code.String()
}

// ErrorMessageForHTTPStatus cubre los errores que no llegan a un handler —una
// ruta que no existe, un cuerpo que no se puede leer— y por eso no pasan por
// el interceptor: sólo queda el estado HTTP para nombrarlos.
func ErrorMessageForHTTPStatus(status int) string {
	switch {
	case status == http.StatusBadRequest:
		return ErrorMessages.InvalidRequestData
	case status == http.StatusUnauthorized:
		return ErrorMessages.Unauthorized
	case status == http.StatusNotFound:
		return ErrorMessages.NotFound
	case status == http.StatusForbidden:
		return connect.CodePermissionDenied.String()
	case status == http.StatusConflict:
		return ErrorMessages.AlreadyExists
	case status == http.StatusPreconditionFailed:
		return connect.CodeFailedPrecondition.String()
	case status == http.StatusTooManyRequests:
		return connect.CodeResourceExhausted.String()
	case status >= 400 && status < 500:
		return ErrorMessages.InvalidRequestData
	}
	return ErrorMessages.InternalServerError
}
