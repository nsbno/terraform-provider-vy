package central_cognito

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	BaseUrl    string
	HTTPClient *http.Client // Optional: if set, used instead of AWS signed requests (for testing)
}

// errorMessageFromResponse returns a readable message for a failed request.
// It uses `detail` from a problem-details body, or `message` from a `{"message", "error_type"}` body.
// Otherwise it returns the raw body, or the status code if the body is empty.
func errorMessageFromResponse(statusCode int, body []byte) string {
	var parsed struct {
		Detail  string `json:"detail"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &parsed); err == nil {
		if parsed.Detail != "" {
			return parsed.Detail
		}

		if parsed.Message != "" {
			return parsed.Message
		}
	}

	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return fmt.Sprintf("unexpected status code %d", statusCode)
	}

	return raw
}
