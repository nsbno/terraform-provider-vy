package central_cognito

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	BaseUrl string
}

// signedRequest sends a request to our endpoint with AWS Signature V4.
// This is how we authenticate with the API.
func signedRequest(request *http.Request) (*http.Response, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	signer := v4.NewSigner(sess.Config.Credentials)

	// We could just pass in the original body, but it feels kinda wasteful API wise.
	var body io.ReadSeeker
	if request.Body != nil {
		reader, _ := ioutil.ReadAll(request.Body)
		body = bytes.NewReader(reader)
	}

	_, err := signer.Sign(request, body, "execute-api", "eu-west-1", time.Now())
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}
