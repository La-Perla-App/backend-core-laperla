package middlewares

import (
	"net/http"
	"strings"
)

// Middleware provides a convenient mechanism for filtering HTTP requests
// entering the application. It returns a new handler which may perform various
// operations and should finish by calling the next HTTP handler.
type Middleware = func(next http.HandlerFunc) http.HandlerFunc

// middlewaresChain provides syntactic sugar to create a new middleware
// which will be the result of chaining the ones received as parameters.
func middlewaresChain(mw ...Middleware) Middleware {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](last)
			}
			last(w, r)
		}
	}
}

func isGETRequest(r *http.Request) bool {
	return isGETRequestByString(r.Method)
}

func isGETRequestByString(method string) bool {
	return method == "" || strings.EqualFold(method, "get")
}

func hasErrorsByMiddleware(r *http.Request) bool {
	return r.Header.Get(HeadersKeys.HasErrors) == "true"
}

func showGrpcErrorDetails(r *http.Request) bool {
	return r.Header.Get(HeadersKeys.ShowGRPCErrorDetails) == "true"
}

func purgeInternalHeaders(r *http.Request) {
	r.Header.Del("Has-Error")
	r.Header.Del("X-Custom-Handler")
	r.Header.Del("X-Encrypted-Prefix")
}

func processNulls(data any) any {
	switch v := data.(type) {
	case map[string]any:
		keys := len(v)
		for key, value := range v {
			if strings.HasPrefix(key, "$ignore") {
				delete(v, key)
				return processNulls(data)
			}
			if keys == 1 {
				switch key {
				case "$null":
					if value == nil {
						return nil
					}
				case "$value":
					return processNulls(value)
				}
			}
			v[key] = processNulls(value)
		}
	case []any:
		// Recorrer el array y procesar cada valor
		for i, item := range v {
			v[i] = processNulls(item)
		}
	}
	return data
}

func isConnectRPCRequest(r *http.Request) bool {
	if r.Header.Get("Connect-Protocol-Version") != "" {
		return true
	}

	contentType := r.Header.Get("Content-Type")
	userAgent := r.Header.Get("User-Agent")

	if strings.HasPrefix(contentType, "application/connect") ||
		strings.HasPrefix(contentType, "application/proto") ||
		strings.HasPrefix(contentType, "application/grpc") {
		return true
	}

	if strings.HasPrefix(userAgent, "connect") ||
		strings.HasPrefix(contentType, "grpc") {
		return true
	}

	return false
}
