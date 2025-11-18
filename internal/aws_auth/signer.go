package aws_auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
)

// SignedRequest sends a request to our endpoint with AWS Signature V4.
// This is how we authenticate with the API.
func SignedRequest(request *http.Request) (*http.Response, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	credentials, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		return nil, err
	}

	signer := v4.NewSigner()

	// Compute the SHA256 hash of the request body for signing
	var payloadHash string
	if request.Body != nil {
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		// Compute SHA256 hash
		hash := sha256.Sum256(bodyBytes)
		payloadHash = hex.EncodeToString(hash[:])
		// Restore the body for the actual request
		request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	} else {
		// Empty body hash (SHA256 of empty string)
		hash := sha256.Sum256([]byte{})
		payloadHash = hex.EncodeToString(hash[:])
	}

	err = signer.SignHTTP(context.Background(), credentials, request, payloadHash, "execute-api", "eu-west-1", time.Now())
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}
