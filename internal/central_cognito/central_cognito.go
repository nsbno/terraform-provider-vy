package central_cognito

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	BaseUrl string
}

type Scope struct {
	Name        string
	Description string
}

type ResourceServer struct {
	Identifier string
	Name       string
	Scopes     []Scope
}

type ResourceServerUpdateRequest struct {
	Identifier string
	Name       string
	Scopes     []Scope
}

// signedRequest sends a request to our endpoint with AWS Signature V4.
// This is how we authenticate with the API.
func signedRequest(request *http.Request) (*http.Response, error) {
	creds := credentials.NewEnvCredentials()
	signer := v4.NewSigner(creds)

	// We could just pass in the original body, but it feels kinda wasteful API wise.
	reader, _ := ioutil.ReadAll(request.Body)
	body := bytes.NewReader(reader)

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
		http.MethodPost,
		fmt.Sprintf("https://%s/resource-servers/%s", c.BaseUrl, identifier),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := signedRequest(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(server)
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
		return errors.New("could not create resource... Returned non-201")
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
		return errors.New("could not delete resource")
	}

	return nil
}
