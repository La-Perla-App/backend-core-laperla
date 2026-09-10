package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/La-Perla-App/backend-core-laperla/pkg/api"
	"github.com/La-Perla-App/backend-core-laperla/pkg/api/auth"
)

var prodMode = os.Getenv("MODE") == "PROD"

var Middlewares = []Middleware{
	DefaultGeneralMiddlewares,
}

var DefaultGeneralMiddlewares = middlewaresChain([]Middleware{
	JSONContentExtractor,
	addGeneralParamsMiddleware,
	wrapJsonResponse,
}...)

func ApplyTranscoderMiddlewares(gatewayPrefixes []string) (Middleware, error) {
	return func(next http.HandlerFunc) http.HandlerFunc {
		handler := func(w http.ResponseWriter, r *http.Request) {
			handlerFn := next

			// Probes de k8s: no pasar por GeneralParams / AUTH WARN.
			switch r.URL.Path {
			case "/health", "/healthz", "/readyz":
				handlerFn.ServeHTTP(w, r)
				return
			}

			if isConnectRPCRequest(r) {
				handlerFn.ServeHTTP(w, r)
				return
			}

			// Purge internal headers
			purgeInternalHeaders(r)

			middlewaresChain(Middlewares...)(handlerFn).ServeHTTP(w, r)
		}
		return handler
	}, nil
}

func addGeneralParamsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		generalParams := api.GeneralParamsFromHTTPRequest(r)
		api.SetGeneralParamsHeader(generalParams, r.Header)
		next.ServeHTTP(w, r)
	})
}

func JSONContentExtractor(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isGETRequest(r) {
			next.ServeHTTP(w, r)
			return
		}
		contentType := r.Header.Get("Content-Type")
		segments := strings.Split(contentType, ";")
		if len(segments) > 1 {
			contentType = ""
		}
		if contentType == "" || strings.HasPrefix(contentType, "text/plain") {
			r.Header.Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	}
}

type jsonResponseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rw *jsonResponseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *jsonResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

func (rw *jsonResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *jsonResponseWriter) Flush() {
}

func wrapJsonResponse(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(HeadersKeys.CustomHandler) != "" {
			r.Header.Del(HeadersKeys.CustomHandler)
			next(w, r)
			return
		}

		generalParams, _ := api.GeneralParamsFromHeaders(r.Header)

		if r.Header.Get("Accept-Encoding") != "" {
			r.Header.Del("Accept-Encoding")
		}

		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed reading request body", http.StatusInternalServerError)
			return
		}

		if !prodMode {
			log.Println("Request: ", string(requestBody))
		}

		if !isGETRequest(r) && strings.TrimSpace(string(requestBody)) == "" {
			requestBody = []byte("{}")
		}
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		r.ContentLength = int64(len(requestBody))

		rw := &jsonResponseWriter{ResponseWriter: w, body: &bytes.Buffer{}}
		next.ServeHTTP(rw, r)

		if rw.statusCode == http.StatusNoContent {
			w.Header().Del("Content-Type")
			w.Header().Del("Content-Length")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		rw.Header().Del("Content-Length")

		w.Header().Del("Content-Encoding")
		w.Header().Del("Accept-Encoding")

		w.Header().Del(api.ResponseInfoHeaderKey)

		contentType := rw.Header().Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			// Read the original response body
			var originalBody any
			err := json.Unmarshal(rw.body.Bytes(), &originalBody)
			if err != nil {
				http.Error(w, "Failed to unmarshal JSON response", http.StatusInternalServerError)
				return
			}
			originalBody = processNulls(originalBody)

			var hasError bool
			if rw.statusCode >= 200 && rw.statusCode < 400 {
				hasError = hasErrorsByMiddleware(r)
			} else {
				hasError = true
			}

			appInfo, _ := api.ResponseInfoFromHeaders(r.Header)
			if hasError {
				appInfo.FillErrorInfo()
			} else {
				appInfo.FillSuccessInfo()
			}

			appInfo.ShowGRPCErrorDetails = showGrpcErrorDetails(r)

			wrappedResponse := ResponseData{
				Message: appInfo.Message,
			}

			if (hasError && appInfo.ShowGRPCErrorDetails) || !hasError {
				wrappedResponse.Data = originalBody
			}

			if appInfo.ResponseContent != nil {
				wrappedResponse.Data = appInfo.ResponseContent
			}

			if wrappedResponse.Data != nil {
				data, _ := json.Marshal(wrappedResponse.Data)
				if string(data) == "{}" {
					wrappedResponse.Data = nil
				}
			}

			w.Header().Set("Content-Type", "application/json")

			if appInfo.SessionToken != "" {
				http.SetCookie(w, auth.CookieFromSessionToken(appInfo.SessionToken))
			}

			clientId := generalParams.ClientId
			if appInfo.ClientId != "" {
				clientId = appInfo.ClientId
			}

			if _, err := r.Cookie(api.ClientIDCookieName); err != nil || appInfo.ClientId != "" {
				if clientId != "" {
					cookie := http.Cookie{
						Name:     api.ClientIDCookieName,
						Value:    clientId,
						HttpOnly: true,
						SameSite: http.SameSiteLaxMode,
						Path:     "/",
					}
					http.SetCookie(w, &cookie)
				}
			}

			statusCode := rw.statusCode
			if appInfo.HTTPStatusCode != 0 {
				statusCode = appInfo.HTTPStatusCode
			} else if hasError && appInfo.APIErrorCode > api.APIErrorCode(connect.CodeResourceExhausted) {
				statusCode = int(appInfo.APIErrorCode)
			}
			w.WriteHeader(statusCode)
			wrappedResponse.StatusCode = statusCode

			jsonResponse, _ := json.Marshal(wrappedResponse)

			if !prodMode {
				log.Println("Response: ", string(jsonResponse))
			}

			w.Write(jsonResponse)
		} else {
			w.WriteHeader(rw.statusCode)
			w.Write(rw.body.Bytes())
		}
	})
}
