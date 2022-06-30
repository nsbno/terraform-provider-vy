package central_cognito

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type AppClient struct {
	Name           string   `json:"name"`
	Scopes         []string `json:"scopes"`
	Type           string   `json:"type"`
	CallbackUrls   []string `json:"callback_urls"`
	LogoutUrls     []string `json:"logout_urls"`
	Secret         string   `json:"secret"`
	GenerateSecret bool     `json:"generate_secret"`
}

type AppClientUpdateRequest struct {
	Name         string   `json:"name"`
	Scopes       []string `json:"scopes"`
	CallbackUrls string   `json:"callback_urls"`
	LogoutUrls   string   `json:"logout_urls"`
}

func (c Client) ReadAppClient(name string, server *AppClient) error {
	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://%s/app-clients/%s", c.BaseUrl, name),
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

func (c Client) CreateAppClient(server AppClient) error {
	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(server)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://%s/app-clients", c.BaseUrl),
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

func (c Client) UpdateAppClient(updateRequest AppClientUpdateRequest) error {
	var data bytes.Buffer

	err := json.NewEncoder(&data).Encode(updateRequest)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("https://%s/app-clients/%s", c.BaseUrl, updateRequest.Name),
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
		return errors.New("could not update resource")
	}

	return nil
}

func (c Client) DeleteAppClient(name string) error {
	request, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("https://%s/app-clients/%s", c.BaseUrl, name),
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
