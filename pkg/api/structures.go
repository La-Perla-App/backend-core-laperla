package api

import (
	"encoding/json"

	"connectrpc.com/connect"
	"github.com/La-Perla-App/backend-core-laperla/pkg/api/auth"
)

type ResponseType string

// ResponseInfo header struct, used to build info section of response
type ResponseInfo struct {
	// Response type: error, success, warning, etc.
	Type ResponseType `json:"type"`

	// HTTP status code respose
	HTTPStatusCode int

	// Custom API error code (non-IANA/connect codes like 427)
	APIErrorCode APIErrorCode `json:"apiErrorCode,omitempty"`

	AvoidSetSuccessfulInfo bool
	AvoidSetErrorInfo      bool

	ClientId string

	// Message descriptor response
	Message string `json:"message,omitempty"`

	// The user session token
	SessionToken string `json:"sessionToken,omitempty"`

	// Response overwrite for HTTP transcoding
	ResponseContent any `json:"responseContent,omitempty"`

	// Show gRPC handler error details on respnse
	ShowGRPCErrorDetails bool `json:"showGRPCErrorDetails,omitempty"`
}

func (info ResponseInfo) Encode() (string, error) {
	jsonInfo, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	return connect.EncodeBinaryHeader(jsonInfo), nil
}

func (appInfo *ResponseInfo) FillErrorInfo() {
	if appInfo.Type == "" {
		appInfo.Type = ResponseTypes.Error
	}
	if appInfo.AvoidSetErrorInfo {
		return
	}
	if appInfo.Message == "" {
		appInfo.Message = ErrorMessages.InternalServerError
	}
}

func (appInfo *ResponseInfo) FillSuccessInfo() {
	if appInfo.Type == "" {
		appInfo.Type = ResponseTypes.Success
	}
	if appInfo.AvoidSetSuccessfulInfo {
		return
	}
	if appInfo.Message == "" {
		appInfo.Message = DefaultSuccessMessage
	}
}

type GeneralParams struct {
	Session       *auth.SessionData
	SessionToken  string
	Lang          string
	ClientId      string
	IANATimezone  string
	Platform      string
	ClientVersion string
}
