package central_cognito

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

func (c Client) ReadResourceServer(identifier string, server *ResourceServer) error {
	resp, err := http.Get(fmt.Sprintf("https://%s/resource-servers/%s", c.BaseUrl, identifier))
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

	response, err := http.Post(
		fmt.Sprintf("https://%s/resource-servers", c.BaseUrl),
		"application/json",
		&data,
	)
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

	response, err := http.DefaultClient.Do(request)
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

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return errors.New("could not delete resource")
	}

	return nil
}
