package central_cognito

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	BaseUrl string
}

type Scope struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ResourceServer struct {
	Identifier string  `json:"identifier"`
	Name       string  `json:"name"`
	Scopes     []Scope `json:"scopes"`
}

type ResourceServerUpdateRequest struct {
	Identifier string  `json:"identifier"`
	Name       string  `json:"name"`
	Scopes     []Scope `json:"scopes"`
}

// signedRequest sends a request to our endpoint with AWS Signature V4.
// This is how we authenticate with the API.
func signedRequest(request *http.Request) (*http.Response, error) {
	creds := credentials.NewSharedCredentials("", "") // Use default options
	signer := v4.NewSigner(creds)

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

func (c Client) ReadResourceServer(identifier string, server *ResourceServer) error {
	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://%s/resource-servers/%s", c.BaseUrl, identifier),
		nil,
	)
	if err != nil {
		return err
	}

	response, err := signedRequest(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		str, _ := io.ReadAll(response.Body)

		return errors.New(fmt.Sprintf("could not read resource. %s", str))

	}

	err = json.NewDecoder(response.Body).Decode(server)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) CreateResourceServer(server ResourceServer) error {
	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(server)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://%s/resource-servers", c.BaseUrl),
		&data,
	)
	if err != nil {
		return err
	}

	response, err := signedRequest(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 201 {
		defer response.Body.Close()

		str, _ := io.ReadAll(response.Body)

		return errors.New(fmt.Sprintf("could not create resource. %s", str))
	}

	return nil
}

func (c Client) UpdateResourceServer(updateRequest ResourceServerUpdateRequest) error {
	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(updateRequest)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("https://%s/resource-servers/%s", c.BaseUrl, updateRequest.Identifier),
		&data,
	)
	if err != nil {
		return err
	}

	response, err := signedRequest(request)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New("could not delete resource")
	}

	return nil
}

func (c Client) DeleteResourceServer(identifier string) error {
	request, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("https://%s/resource-servers/%s", c.BaseUrl, identifier),
		nil,
	)
	if err != nil {
		return err
	}

	response, err := signedRequest(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		defer response.Body.Close()

		str, _ := io.ReadAll(response.Body)

		return errors.New(fmt.Sprintf("could not delete resource. %s", str))
	}

	return nil
}
