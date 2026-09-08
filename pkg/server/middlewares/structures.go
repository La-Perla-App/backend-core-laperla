package middlewares

type ResponseData struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message,omitempty"`
	Data       any    `json:"data,omitempty"`
}
