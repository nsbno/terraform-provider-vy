package version_handler_v2

import "net/http"

type Client struct {
	BaseUrl    string
	HTTPClient *http.Client // Optional: if set, used instead of AWS signed requests (for testing)
}

type apiErrorPayload struct {
	Message   string `json:"message"`
	ErrorType string `json:"error_type"`
}
