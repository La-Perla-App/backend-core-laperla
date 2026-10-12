package api

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"connectrpc.com/connect"
)

func TestErrorMessageFor(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		// El caso que lo destapó: falta destinationId → 400 que decía
		// internal_server_error.
		{connect.NewError(connect.CodeInvalidArgument, errors.New("destination_id: required")), "invalid_request_data"},
		{connect.NewError(connect.CodeNotFound, errors.New("x")), "not_found"},
		{connect.NewError(connect.CodeAlreadyExists, errors.New("x")), "already_exists"},
		{connect.NewError(connect.CodeUnauthenticated, errors.New("x")), "invalid_token"},
		// Fuera del catálogo: el nombre del código, no «se rompió algo».
		{connect.NewError(connect.CodePermissionDenied, errors.New("x")), "permission_denied"},
		{connect.NewError(connect.CodeFailedPrecondition, errors.New("x")), "failed_precondition"},
		// Envuelto también cuenta.
		{fmt.Errorf("list: %w", connect.NewError(connect.CodeNotFound, errors.New("x"))), "not_found"},
		// Lo que sí es un fallo sigue siéndolo.
		{connect.NewError(connect.CodeInternal, errors.New("x")), "internal_server_error"},
		{connect.NewError(connect.CodeUnknown, errors.New("x")), "internal_server_error"},
		{errors.New("pq: connection refused"), "internal_server_error"},
	}
	for _, c := range cases {
		if got := ErrorMessageFor(c.err); got != c.want {
			t.Errorf("ErrorMessageFor(%v) = %q, quiero %q", c.err, got, c.want)
		}
	}
}

func TestErrorMessageForHTTPStatus(t *testing.T) {
	cases := map[int]string{
		http.StatusBadRequest:          "invalid_request_data",
		http.StatusUnauthorized:        "invalid_token",
		http.StatusForbidden:           "permission_denied",
		http.StatusNotFound:            "not_found",
		http.StatusConflict:            "already_exists",
		http.StatusTooManyRequests:     "resource_exhausted",
		http.StatusUnprocessableEntity: "invalid_request_data",
		http.StatusInternalServerError: "internal_server_error",
		http.StatusBadGateway:          "internal_server_error",
	}
	for status, want := range cases {
		if got := ErrorMessageForHTTPStatus(status); got != want {
			t.Errorf("ErrorMessageForHTTPStatus(%d) = %q, quiero %q", status, got, want)
		}
	}
}
