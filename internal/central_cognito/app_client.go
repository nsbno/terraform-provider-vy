package central_cognito

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/nsbno/terraform-provider-vy/internal/aws_auth"
)

type AppClient struct {
	Name           string   `json:"name"`
	Scopes         []string `json:"scopes"`
	Type           string   `json:"type"`
	GenerateSecret *bool    `json:"generate_secret"`
	CallbackUrls   []string `json:"callback_urls"`
	LogoutUrls     []string `json:"logout_urls"`
	ClientId       *string  `json:"client_id"`
	ClientSecret   *string  `json:"client_secret"`
}

type AppClientUpdateRequest struct {
	Name         string   `json:"name"`
	Scopes       []string `json:"scopes"`
	CallbackUrls []string `json:"callback_urls"`
	LogoutUrls   []string `json:"logout_urls"`
}

func (c Client) ReadAppClient(name string, server *AppClient) error {
	protocol := "https://"
	if c.HTTPClient != nil {
		protocol = "http://"
	}

	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s%s/app-clients/%s", protocol, c.BaseUrl, url.QueryEscape(name)),
		nil,
	)
	if err != nil {
		return err
	}

	var response *http.Response
	if c.HTTPClient != nil {
		response, err = c.HTTPClient.Do(request)
	} else {
		response, err = aws_auth.SignedRequest(request)
	}
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

func (c Client) CreateAppClient(server AppClient) (*AppClient, error) {
	protocol := "https://"
	if c.HTTPClient != nil {
		protocol = "http://"
	}

	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(server)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s%s/app-clients", protocol, c.BaseUrl),
		&data,
	)
	if err != nil {
		return nil, err
	}

	var response *http.Response
	if c.HTTPClient != nil {
		response, err = c.HTTPClient.Do(request)
	} else {
		response, err = aws_auth.SignedRequest(request)
	}
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 201 {
		defer response.Body.Close()

		str, _ := io.ReadAll(response.Body)

		return nil, errors.New(fmt.Sprintf("could not create resource. %s", str))
	}

	var createdAppClient AppClient
	err = json.NewDecoder(response.Body).Decode(&createdAppClient)
	if err != nil {
		return nil, err
	}

	return &createdAppClient, nil
}

func (c Client) UpdateAppClient(updateRequest AppClientUpdateRequest) error {
	protocol := "https://"
	if c.HTTPClient != nil {
		protocol = "http://"
	}

	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(updateRequest)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s%s/app-clients/%s", protocol, c.BaseUrl, url.QueryEscape(updateRequest.Name)),
		&data,
	)
	if err != nil {
		return err
	}

	var response *http.Response
	if c.HTTPClient != nil {
		response, err = c.HTTPClient.Do(request)
	} else {
		response, err = aws_auth.SignedRequest(request)
	}
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

func (c Client) DeleteAppClient(name string) error {
	protocol := "https://"
	if c.HTTPClient != nil {
		protocol = "http://"
	}

	request, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s%s/app-clients/%s", protocol, c.BaseUrl, url.QueryEscape(name)),
		nil,
	)
	if err != nil {
		return err
	}

	var response *http.Response
	if c.HTTPClient != nil {
		response, err = c.HTTPClient.Do(request)
	} else {
		response, err = aws_auth.SignedRequest(request)
	}
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

type ImportAppClientRequest struct {
	ClientId string `json:"client_id"`
}

func (c Client) ImportAppClient(client_id string, server *AppClient) error {
	protocol := "https://"
	if c.HTTPClient != nil {
		protocol = "http://"
	}

	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(ImportAppClientRequest{client_id})
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s%s/import/app-client", protocol, c.BaseUrl),
		&data,
	)
	if err != nil {
		return err
	}

	var response *http.Response
	if c.HTTPClient != nil {
		response, err = c.HTTPClient.Do(request)
	} else {
		response, err = aws_auth.SignedRequest(request)
	}
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
