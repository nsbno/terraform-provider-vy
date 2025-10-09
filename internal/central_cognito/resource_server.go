package central_cognito

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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

func (c Client) ReadResourceServer(identifier string, server *ResourceServer) error {
	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://%s/resource-servers/%s", c.BaseUrl, url.QueryEscape(identifier)),
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
		fmt.Sprintf("https://%s/resource-servers/%s", c.BaseUrl, url.QueryEscape(updateRequest.Identifier)),
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
		defer response.Body.Close()

		str, _ := io.ReadAll(response.Body)

		return errors.New(fmt.Sprintf("could not update resource. %s", str))
	}

	return nil
}

func (c Client) DeleteResourceServer(identifier string) error {
	request, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("https://%s/resource-servers/%s", c.BaseUrl, url.QueryEscape(identifier)),
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

type ImportResourceServerRequest struct {
	Identifier string `json:"identifier"`
}

func (c Client) ImportResourceServer(identifier string, server *ResourceServer) error {
	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(ImportResourceServerRequest{identifier})
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://%s/import/resource-server", c.BaseUrl),
		&data,
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

		return errors.New(fmt.Sprintf("could not import resource. %s", str))

	}

	err = json.NewDecoder(response.Body).Decode(server)
	if err != nil {
		return err
	}

	return nil
}
