package central_cognito

import "net/http"

type Client struct {
	BaseUrl    string
	HTTPClient *http.Client // Optional: if set, used instead of AWS signed requests (for testing)
}
